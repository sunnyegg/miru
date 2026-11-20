package anilist

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// AniList enum values accepted by AdvancedSearch. The app exposes these lists
// through AnimeFilterOptions so the sheet never offers a value the API rejects.
var (
	MediaGenres = []string{
		"Action", "Adventure", "Comedy", "Drama", "Ecchi", "Fantasy", "Horror",
		"Mahou Shoujo", "Mecha", "Music", "Mystery", "Psychological", "Romance",
		"Sci-Fi", "Slice of Life", "Sports", "Supernatural", "Thriller",
	}
	MediaFormats  = []string{"TV", "TV_SHORT", "MOVIE", "SPECIAL", "OVA", "ONA", "MUSIC"}
	MediaStatuses = []string{"FINISHED", "RELEASING", "NOT_YET_RELEASED", "CANCELLED", "HIATUS"}
	MediaSorts    = []string{"POPULARITY_DESC", "SCORE_DESC", "TRENDING_DESC", "START_DATE_DESC", "TITLE_ROMAJI"}
)

const (
	advancedSearchPageSize = 50
	advancedSearchMinYear  = 1940
	advancedSearchMaxQuery = 200
)

// AdvancedSearch holds the filters of one AniList advanced search request.
// Every field is optional, but at least one filter must be set.
type AdvancedSearch struct {
	Search        string
	Genres        []string
	ExcludeGenres []string
	Year          int
	Season        string
	Formats       []string
	Statuses      []string
	MinScore      int
	Sort          string
	Page          int
}

// NormalizeAdvancedSearch trims the input, maps enum values to their AniList
// spelling, and rejects anything the API would refuse.
func NormalizeAdvancedSearch(f AdvancedSearch) (AdvancedSearch, error) {
	f.Search = strings.TrimSpace(f.Search)
	if len(f.Search) > advancedSearchMaxQuery {
		return f, fmt.Errorf("anime search title is too long (max %d characters)", advancedSearchMaxQuery)
	}

	var err error
	if f.Genres, err = normalizeGenreList(f.Genres); err != nil {
		return f, err
	}
	if f.ExcludeGenres, err = normalizeGenreList(f.ExcludeGenres); err != nil {
		return f, err
	}
	if f.Formats, err = normalizeEnumList(f.Formats, MediaFormats, "format"); err != nil {
		return f, err
	}
	if f.Statuses, err = normalizeEnumList(f.Statuses, MediaStatuses, "status"); err != nil {
		return f, err
	}

	f.Season = strings.ToUpper(strings.TrimSpace(f.Season))
	if f.Season != "" && !ValidMediaSeason(f.Season) {
		return f, fmt.Errorf("unknown season %q; valid seasons: WINTER, SPRING, SUMMER, FALL", f.Season)
	}
	maxYear := time.Now().Year() + 2
	if f.Year != 0 && (f.Year < advancedSearchMinYear || f.Year > maxYear) {
		return f, fmt.Errorf("year must be between %d and %d", advancedSearchMinYear, maxYear)
	}
	if f.Season != "" && f.Year == 0 {
		return f, errors.New("choose a year together with a season")
	}
	if f.MinScore < 0 || f.MinScore > 100 {
		return f, errors.New("minimum score must be between 0 and 100")
	}

	f.Sort = strings.ToUpper(strings.TrimSpace(f.Sort))
	if f.Sort == "" {
		f.Sort = MediaSorts[0]
	}
	if _, err := normalizeEnumList([]string{f.Sort}, MediaSorts, "sort"); err != nil {
		return f, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	return f, nil
}

// IsEmpty reports whether the request carries no filter at all.
func (f AdvancedSearch) IsEmpty() bool {
	return f.Search == "" &&
		len(f.Genres) == 0 &&
		len(f.ExcludeGenres) == 0 &&
		f.Year == 0 &&
		f.Season == "" &&
		len(f.Formats) == 0 &&
		len(f.Statuses) == 0 &&
		f.MinScore == 0
}

// SearchAdvanced runs an advanced AniList search and returns one page.
func (c *Client) SearchAdvanced(f AdvancedSearch) (SeasonPage, error) {
	f, err := NormalizeAdvancedSearch(f)
	if err != nil {
		return SeasonPage{}, err
	}
	if f.IsEmpty() {
		return SeasonPage{}, errors.New("advanced anime search needs at least one filter")
	}

	const q = `
	query ($page: Int, $search: String, $genres: [String], $excludeGenres: [String], $season: MediaSeason, $seasonYear: Int, $formats: [MediaFormat], $statuses: [MediaStatus], $minScore: Int, $sort: [MediaSort]) {
	  Page(page: $page, perPage: 50) {
	    pageInfo { hasNextPage }
	    media(type: ANIME, search: $search, genre_in: $genres, genre_not_in: $excludeGenres, season: $season, seasonYear: $seasonYear, format_in: $formats, status_in: $statuses, averageScore_greater: $minScore, sort: $sort) {
	      id
	      title { romaji english }
	      coverImage { large }
	      episodes
	      status
	      season
	      seasonYear
	      description(asHtml: false)
	      mediaListEntry { status }
	    }
	  }
	}`

	variables := map[string]any{
		"page": f.Page,
		"sort": []string{f.Sort},
	}
	if f.Search != "" {
		variables["search"] = f.Search
	}
	if len(f.Genres) > 0 {
		variables["genres"] = f.Genres
	}
	if len(f.ExcludeGenres) > 0 {
		variables["excludeGenres"] = f.ExcludeGenres
	}
	if f.Season != "" {
		variables["season"] = f.Season
	}
	if f.Year != 0 {
		variables["seasonYear"] = f.Year
	}
	if len(f.Formats) > 0 {
		variables["formats"] = f.Formats
	}
	if len(f.Statuses) > 0 {
		variables["statuses"] = f.Statuses
	}
	if f.MinScore > 0 {
		// averageScore_greater is exclusive, so ask for one point below the
		// chosen minimum to keep that score itself in the results.
		variables["minScore"] = f.MinScore - 1
	}

	var out struct {
		Page struct {
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
			Media []gqlMedia `json:"media"`
		} `json:"Page"`
	}
	if err := c.query(q, variables, &out); err != nil {
		return SeasonPage{}, err
	}
	animes := make([]Anime, 0, len(out.Page.Media))
	for _, media := range out.Page.Media {
		animes = append(animes, media.toAnime())
	}
	return SeasonPage{
		Items:       animes,
		HasNextPage: out.Page.PageInfo.HasNextPage,
	}, nil
}

func normalizeGenreList(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		matched := ""
		for _, genre := range MediaGenres {
			if strings.EqualFold(genre, value) {
				matched = genre
				break
			}
		}
		if matched == "" {
			return nil, fmt.Errorf("unknown genre %q; valid genres: %s", value, strings.Join(MediaGenres, ", "))
		}
		out = append(out, matched)
	}
	return out, nil
}

func normalizeEnumList(values, valid []string, label string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		matched := ""
		for _, candidate := range valid {
			if candidate == value {
				matched = candidate
				break
			}
		}
		if matched == "" {
			return nil, fmt.Errorf("unknown %s %q; valid values: %s", label, value, strings.Join(valid, ", "))
		}
		out = append(out, matched)
	}
	return out, nil
}
