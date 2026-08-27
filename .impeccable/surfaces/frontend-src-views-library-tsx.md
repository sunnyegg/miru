---
version: 1
slug: "frontend-src-views-library-tsx"
primary_target: "frontend/src/views/Library.tsx"
related_targets: ["frontend/src/App.tsx","frontend/src/components/Sidebar.tsx","frontend/src/style.css"]
---

# Library

**Mode:** Operate

**Audience / job:** Linux anime viewers open Library to browse by show, pick an episode, and Play in MPV.

**Action:** Select a title, then Play. Import file and AniList match stay secondary, inline.

**Direction:** MPV OSC Overlay (seed 7ad0351c, model-pick). Matte black bezel, poster field, bottom OSC strip. Left rail of named destinations. Play is the orange hit on the selected show.

**Memorable moment:** Choosing a poster drops the OSC; episode ticks act as chapters; Play sits on that strip.

**Scope:** Production-ready shell + Library. Other tabs inherit tokens and chrome only.

**Constraints:** No Go/backend changes. Group episodes in the UI. Do not hide Play, drop match, or feel like a website.

**Open:** Grouping is client-side on the flat episode list.

**Approved critique reference:** `.impeccable/mocks/decision/model-pick.webp` (code-led; not a pixel contract)
