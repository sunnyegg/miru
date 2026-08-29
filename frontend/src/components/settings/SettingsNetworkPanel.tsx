import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {SettingsView} from '../../lib/types'
import {Button} from '@/components/ui/button'
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

type Props = {
  form: SettingsView
  setForm: Dispatch<SetStateAction<SettingsView>>
  saving: boolean
  testingNetwork: boolean
  onSubmit: (event: FormEvent) => void
  onTestNetwork: () => void
}

export function SettingsNetworkPanel({
  form,
  setForm,
  saving,
  testingNetwork,
  onSubmit,
  onTestNetwork,
}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>Network</CardTitle>
          <CardDescription>
            Choose how Miru connects to AniList, Nyaa, and downloads.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Label htmlFor="networkMode" className="mb-2">
            Connection mode
          </Label>
          <NativeSelect
            id="networkMode"
            value={form.networkMode}
            onChange={(event) =>
              setForm((current) => ({...current, networkMode: event.target.value}))
            }
          >
            <NativeSelectOption value="system">System proxy / VPN</NativeSelectOption>
            <NativeSelectOption value="direct">Direct connection</NativeSelectOption>
            <NativeSelectOption value="socks5">SOCKS5 proxy</NativeSelectOption>
            <NativeSelectOption value="http_proxy">HTTP/HTTPS proxy</NativeSelectOption>
          </NativeSelect>
          {form.networkMode === 'socks5' && (
            <>
              <Label htmlFor="socks5Address" className="mt-4 mb-2">
                SOCKS5 address
              </Label>
              <Input
                id="socks5Address"
                value={form.socks5Address}
                onChange={(event) =>
                  setForm((current) => ({...current, socks5Address: event.target.value}))
                }
                placeholder="127.0.0.1:1080"
                className="bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Torrent traffic uses TCP through this proxy. UDP, DHT, and inbound peers are
                disabled.
              </p>
            </>
          )}
          {form.networkMode === 'http_proxy' && (
            <>
              <Label htmlFor="httpProxyUrl" className="mt-4 mb-2">
                Proxy URL
              </Label>
              <Input
                id="httpProxyUrl"
                value={form.httpProxyUrl}
                onChange={(event) =>
                  setForm((current) => ({...current, httpProxyUrl: event.target.value}))
                }
                placeholder="http://127.0.0.1:8080"
                className="bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                HTTP and HTTPS traffic routes through this proxy. Use http:// or https:// with host
                and port.
              </p>
            </>
          )}
        </CardContent>
        <CardFooter className="flex flex-wrap gap-2">
          <Button type="submit" variant="secondary" disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
          <Button type="button" variant="muted" disabled={testingNetwork} onClick={onTestNetwork}>
            {testingNetwork ? 'Testing…' : 'Test connection'}
          </Button>
        </CardFooter>
      </Card>
    </form>
  )
}
