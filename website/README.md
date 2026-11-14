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

Release metadata and assets are loaded from the GitHub Releases API at
`https://api.github.com/repos/Grizak/Wick/releases?per_page=100`. Downloads link
directly to each asset's `browser_download_url`; release pages link to the
corresponding `html_url`.

## Development

```sh
pnpm install
pnpm dev
```

Run `pnpm build` to type-check and create a production build, and `pnpm lint` to
run Oxlint.
