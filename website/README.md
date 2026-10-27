# Wick website

The website is a React + TypeScript application built with Vite.

## Data sources

The changelog view reads from the site origin:

- `/changelog.index.json` lists changelog files or entries.
- `/changelog` may return a combined Markdown or JSON changelog as a fallback.
- File entries in the index are loaded from `/changelog/<file>`.

The index can be a string array of filenames, or an object with a
`changelogs`, `entries`, `files`, or `items` array. Entries may be filenames or
objects with `path` (or `file`/`name`), `title`, `version`, `date`, and
`content`/`body` fields. If an entry includes its content, no separate file
request is made.

Release metadata is loaded from
`https://pkg.isaksweb.xyz/index.json`. The `wick` package's versions are read
from that index, then each version record is loaded from
`https://pkg.isaksweb.xyz/packages/wick/<version>.json`. Artifact `path` values
in those records are resolved against the registry host.

## Development

```sh
pnpm install
pnpm dev
```

Run `pnpm build` to type-check and create a production build, and `pnpm lint` to
run Oxlint.
