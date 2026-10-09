# name-generator

A fantasy name generator deployed as the `eagle0/names` DigitalOcean Function.

## Requests

Send a batch with `Accept: application/json`. The response remains
`{"names":[{"id":"...","name":"..."}]}` in request order.

```json
{
  "requests": [
    {"id": "h1", "gender": "female"},
    {"id": "h2", "kind": "hero", "gender": "male"},
    {
      "id": "b1",
      "kind": "battalion",
      "battalionType": "heavy_infantry",
      "provinceName": "Westmarch"
    }
  ]
}
```

Omitting `kind` preserves hero generation. An empty request batch still returns
20 hero names. Battalion requests require a nonblank `provinceName` and one of
`light_infantry`, `heavy_infantry`, `light_cavalry`, `heavy_cavalry`, `longbowmen`,
or `undead`. Gender is ignored for battalions; leader-name fragments can use
names of any gender. Unknown kinds/types and missing provinces return HTTP 400.
HTML output is also supported through the existing content negotiation.

## Configuration and deployment

Both templates and the shared word lists are fetched from the `eagle0-config`
DigitalOcean Spaces bucket and initialized once per function instance:

- `names.tsv`: shared word lists, including battalion types and adjectives.
- `nameConstruction.txt`: hero name template.
- `battalionNameConstruction.txt`: battalion name template.

The files beside `names.go` are checked-in copies for reference and tests;
production loads them from Spaces rather than embedding them. Upload
`packages/eagle0/names/battalionNameConstruction.txt` to that bucket under the
key `battalionNameConstruction.txt` **before deploying this change**. Existing
`names.tsv` must contain all six battalion type lists and their `_adj` lists.
The shared `noun`, `pluralnoun`, `adjective`, `verb`, `name`, and `surname`
lists are also used. The `update-words` function updates `names.tsv` only.

## Tests

Run `go test ./...` from `packages/eagle0/names`. Handler tests use checked-in
configuration without network access. The Spaces integration test runs when
`DIGITALOCEAN_ACCESS_KEY_ID` and `DIGITALOCEAN_SECRET_KEY` are set, otherwise
it is skipped. CI supplies those credentials.
