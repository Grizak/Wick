import { useEffect, useMemo, useRef, useState } from 'react'
import './App.css'

const REGISTRY_URL = 'https://pkg.isaksweb.xyz'

type PackageFile = {
  installcommand?: string
  path?: string
  size?: number
  sha256?: string
}

type PackageVersion = {
  name: string
  version: string
  pkgs?: Record<string, PackageFile[]>
  description?: string
}

type ChangeEntry = {
  id: string
  title: string
  version?: string
  date?: string
  content: string
}

type ChangeIndexItem =
  | string
  | {
      name?: string
      path?: string
      file?: string
      title?: string
      version?: string
      date?: string
      content?: string
      body?: string
    }

type RegistryIndex = {
  packages?: Array<{ name?: string; versions?: string[] }>
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function getIndexItems(value: unknown): ChangeIndexItem[] {
  if (Array.isArray(value)) {
    return value.filter(
      (item): item is ChangeIndexItem =>
        typeof item === 'string' || isRecord(item),
    ) as ChangeIndexItem[]
  }
  if (!isRecord(value)) return []

  for (const key of ['changelogs', 'entries', 'files', 'items']) {
    if (Array.isArray(value[key])) return getIndexItems(value[key])
  }

  return Object.entries(value).map(([name, item]) => {
    if (typeof item === 'string') return { name, content: item }
    if (isRecord(item)) return { ...item, name: String(item.name ?? name) }
    return { name }
  })
}

function getItemString(item: ChangeIndexItem, key: string): string | undefined {
  if (typeof item === 'string') return undefined
  const value = item[key as keyof typeof item]
  return typeof value === 'string' ? value : undefined
}

function getItemPath(item: ChangeIndexItem): string | undefined {
  if (typeof item === 'string') return item
  return item.path ?? item.file ?? item.name
}

function changelogPath(path: string): string {
  const normalized = path.trim().replace(/^\/+/, '')
  if (!normalized || normalized === 'changelog') return '/changelog'
  if (normalized.startsWith('changelog/')) return `/${normalized}`

  const safePath = normalized
    .split('/')
    .filter((part) => part && part !== '.' && part !== '..')
    .map(encodeURIComponent)
    .join('/')
  return `/changelog/${safePath}`
}

function isHtmlDocument(content: string): boolean {
  return /^<(?:!doctype\s+html|html\b)/i.test(content.trimStart())
}

function titleFromMarkdown(content: string): string {
  const heading = content.match(/^\s*#\s+(.+)$/m)
  return heading?.[1]?.trim() ?? 'Release notes'
}

function withoutTitleHeading(content: string, title: string): string {
  const lines = content.replace(/\r\n/g, '\n').split('\n')
  const headingIndex = lines.findIndex((line) => /^\s*#\s+/.test(line))
  if (headingIndex < 0 || lines[headingIndex].replace(/^\s*#\s+/, '').trim() !== title) {
    return content
  }
  lines.splice(headingIndex, 1)
  return lines.join('\n').trimStart()
}

function versionFromTitle(title: string): string | undefined {
  return title.match(/\bv?\d+\.\d+(?:\.\d+)?(?:[-+][\w.-]+)?\b/)?.[0]
}

function normalizeVersion(version: string): string {
  return version.replace(/^v/, '')
}

function compareVersions(left: string, right: string): number {
  const leftParts = normalizeVersion(left).split(/[.+-]/)
  const rightParts = normalizeVersion(right).split(/[.+-]/)
  const length = Math.max(leftParts.length, rightParts.length)

  for (let index = 0; index < length; index += 1) {
    const leftPart = leftParts[index] ?? '0'
    const rightPart = rightParts[index] ?? '0'
    const leftNumber = Number(leftPart)
    const rightNumber = Number(rightPart)
    const comparison =
      Number.isNaN(leftNumber) || Number.isNaN(rightNumber)
        ? leftPart.localeCompare(rightPart)
        : leftNumber - rightNumber
    if (comparison !== 0) return comparison
  }

  return 0
}

function formatDate(value?: string): string | undefined {
  if (!value) return undefined
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('en', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  }).format(date)
}

function formatSize(bytes?: number): string | undefined {
  if (typeof bytes !== 'number' || bytes <= 0) return undefined
  return `${(bytes / 1_000_000).toFixed(1)} MB`
}

async function fetchJson(url: string): Promise<unknown> {
  const response = await fetch(url)
  if (!response.ok) throw new Error(`Request failed (${response.status})`)
  return response.json() as Promise<unknown>
}

async function fetchChangelogFile(item: ChangeIndexItem): Promise<ChangeEntry> {
  const content = getItemString(item, 'content') ?? getItemString(item, 'body')
  const path = getItemPath(item)
  if (content === undefined && !path) {
    throw new Error('A changelog entry is missing its file path or content.')
  }
  const responseText =
    content ??
    (path
      ? await (async () => {
          const response = await fetch(changelogPath(path))
          if (!response.ok) {
            throw new Error(`${path}: request failed (${response.status})`)
          }
          const text = await response.text()
          if (isHtmlDocument(text)) {
            throw new Error(`${path}: expected changelog content, received HTML`)
          }
          return text
        })()
      : '')
  const title = getItemString(item, 'title') ?? titleFromMarkdown(responseText)

  return {
    id: path ?? title,
    title,
    version: getItemString(item, 'version') ?? versionFromTitle(title),
    date: getItemString(item, 'date'),
    content: withoutTitleHeading(responseText.replace(/^\uFEFF/, ''), title),
  }
}

async function loadChangelogs(): Promise<ChangeEntry[]> {
  const [indexResult, directResult] = await Promise.allSettled([
    fetchJson('/changelog.index.json'),
    fetch('/changelog'),
  ])
  let directEntry: ChangeEntry | undefined

  if (directResult.status === 'fulfilled' && directResult.value.ok) {
    const response = directResult.value
    const text = await response.text()
    if (!isHtmlDocument(text)) {
      const contentType = response.headers.get('content-type') ?? ''
      if (contentType.includes('application/json')) {
        try {
          const directItems = getIndexItems(JSON.parse(text) as unknown)
          if (directItems.length > 0) {
            const entries = await Promise.all(directItems.map(fetchChangelogFile))
            if (entries.length > 0) return entries
          }
        } catch {
          directEntry = undefined
        }
      } else if (text.trim()) {
        directEntry = {
          id: 'changelog',
          title: titleFromMarkdown(text),
          version: versionFromTitle(titleFromMarkdown(text)),
          content: withoutTitleHeading(text, titleFromMarkdown(text)),
        }
      }
    }
  }

  if (indexResult.status === 'fulfilled') {
    const items = getIndexItems(indexResult.value)
    if (items.length > 0) {
      const entries = await Promise.all(items.map(fetchChangelogFile))
      return entries.sort((left, right) => {
        if (left.date && right.date) {
          return Date.parse(right.date) - Date.parse(left.date)
        }
        return (right.version ?? '').localeCompare(left.version ?? '', undefined, {
          numeric: true,
        })
      })
    }
  }

  if (directEntry) return [directEntry]
  if (indexResult.status === 'rejected') throw indexResult.reason
  throw new Error('No changelog entries were found.')
}

async function loadWickVersions(): Promise<PackageVersion[]> {
  const index = (await fetchJson(`${REGISTRY_URL}/index.json`)) as RegistryIndex
  const wick = index.packages?.find((item) => item.name === 'wick')
  if (!wick?.versions?.length) throw new Error('Wick versions were not found.')

  const versions = [...wick.versions].sort(compareVersions).reverse()
  return Promise.all(
    versions.map(async (version) => {
      const record = (await fetchJson(
        `${REGISTRY_URL}/packages/wick/${encodeURIComponent(version)}.json`,
      )) as PackageVersion
      if (!record.version || !record.pkgs) {
        throw new Error(`The package record for Wick ${version} is incomplete.`)
      }
      return record
    }),
  )
}

function InlineMarkdown({ text }: { text: string }) {
  const tokens = text.split(/(`[^`]+`)/g)
  return (
    <>
      {tokens.map((token, index) =>
        token.startsWith('`') && token.endsWith('`') ? (
          <code key={`${index}-${token}`}>{token.slice(1, -1)}</code>
        ) : (
          <span key={`${index}-${token}`}>{token}</span>
        ),
      )}
    </>
  )
}

function MarkdownContent({ content }: { content: string }) {
  const lines = content.replace(/\r\n/g, '\n').split('\n')
  const blocks: Array<{ type: string; text: string[] }> = []
  let inCode = false

  for (const line of lines) {
    if (line.trimStart().startsWith('```')) {
      inCode = !inCode
      continue
    }
    if (inCode) {
      const last = blocks.at(-1)
      if (last?.type === 'code') last.text.push(line)
      else blocks.push({ type: 'code', text: [line] })
      continue
    }

    const trimmed = line.trim()
    if (!trimmed) continue
    const heading = trimmed.match(/^(#{1,4})\s+(.+)$/)
    const bullet = trimmed.match(/^[-*]\s+(.+)$/)
    const type = heading ? `h${heading[1].length}` : bullet ? 'li' : 'p'
    const text = heading?.[2] ?? bullet?.[1] ?? trimmed
    const previous = blocks.at(-1)
    if (type === 'li' && previous?.type === 'ul') previous.text.push(text)
    else blocks.push({ type: type === 'li' ? 'ul' : type, text: [text] })
  }

  return (
    <div className="markdown-content">
      {blocks.map((block, index) => {
        const key = `${block.type}-${index}`
        if (block.type === 'ul') {
          return (
            <ul key={key}>
              {block.text.map((line, lineIndex) => (
                <li key={`${key}-${lineIndex}`}>
                  <InlineMarkdown text={line} />
                </li>
              ))}
            </ul>
          )
        }
        if (block.type === 'code') {
          return (
            <pre key={key}>
              <code>{block.text.join('\n')}</code>
            </pre>
          )
        }
        if (/^h[1-4]$/.test(block.type)) {
          const Heading = block.type as 'h1' | 'h2' | 'h3' | 'h4'
          return (
            <Heading key={key}>
              <InlineMarkdown text={block.text[0]} />
            </Heading>
          )
        }
        return (
          <p key={key}>
            <InlineMarkdown text={block.text[0]} />
          </p>
        )
      })}
    </div>
  )
}

function PackageDownloads({ release }: { release: PackageVersion }) {
  const groups = Object.entries(release.pkgs ?? {}).filter(
    ([, files]) => files.length > 0,
  )

  if (groups.length === 0) {
    return <p className="empty-note">No downloadable builds listed for this version.</p>
  }

  return (
    <>
      <a href={`https://github.com/Grizak/Wick/releases/tag/${release}`} className="download-link" rel="noreferrer" target='_blank'>Open on github releases</a>
      <div className="download-grid">
        {groups.map(([platform, files]) =>
          files.map((file, index) => (
            <a
              className="download-link"
              href={file.path ? `${REGISTRY_URL}/${file.path.replace(/^\/+/, '')}` : undefined}
              key={`${platform}-${file.path ?? index}`}
              rel="noreferrer"
              target="_blank"
            >
              <span className="download-icon" aria-hidden="true">↓</span>
              <span className="download-copy">
                <strong>{platform}</strong>
                <small>{file.path?.split('/').at(-1) ?? 'Package download'}</small>
                {file.installcommand && (
                  <code className="download-command">{file.installcommand}</code>
                )}
              </span>
              <span className="download-size">{formatSize(file.size)}</span>
            </a>
          )),
        )}
      </div>
    </>
  )
}

function App() {
  const searchInput = useRef<HTMLInputElement>(null)
  const [changelogs, setChangelogs] = useState<ChangeEntry[]>([])
  const [versions, setVersions] = useState<PackageVersion[]>([])
  const [changelogError, setChangelogError] = useState('')
  const [registryError, setRegistryError] = useState('')
  const [loadingChangelogs, setLoadingChangelogs] = useState(true)
  const [loadingRegistry, setLoadingRegistry] = useState(true)
  const [selectedVersion, setSelectedVersion] = useState('')
  const [search, setSearch] = useState('')

  useEffect(() => {
    let active = true
    loadChangelogs()
      .then((entries) => {
        if (active) setChangelogs(entries)
      })
      .catch((error: unknown) => {
        if (active) {
          setChangelogError(
            error instanceof Error ? error.message : 'Unable to load changelogs.',
          )
        }
      })
      .finally(() => {
        if (active) setLoadingChangelogs(false)
      })

    loadWickVersions()
      .then((releases) => {
        if (active) {
          setVersions(releases)
          setSelectedVersion(releases[0]?.version ?? '')
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setRegistryError(
            error instanceof Error ? error.message : 'Unable to load releases.',
          )
        }
      })
      .finally(() => {
        if (active) setLoadingRegistry(false)
      })

    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    function focusSearch(event: KeyboardEvent) {
      if (
        event.key === '/' &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.altKey &&
        !(event.target instanceof HTMLInputElement) &&
        !(event.target instanceof HTMLTextAreaElement) &&
        !(event.target instanceof HTMLElement && event.target.isContentEditable)
      ) {
        event.preventDefault()
        searchInput.current?.focus()
      }
    }

    window.addEventListener('keydown', focusSearch)
    return () => window.removeEventListener('keydown', focusSearch)
  }, [])

  const filteredChangelogs = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return changelogs
    return changelogs.filter((entry) =>
      `${entry.title} ${entry.version ?? ''} ${entry.content}`
        .toLowerCase()
        .includes(query),
    )
  }, [changelogs, search])

  const selectedRelease = versions.find(
    (release) => release.version === selectedVersion,
  )

  return (
    <div className="site-shell">
      <header className="topbar">
        <a className="brand" href="#" aria-label="Wick home">
          <span className="brand-mark" aria-hidden="true">w</span>
          <span>wick<span className="brand-period">.</span></span>
        </a>
        <nav className="nav-links" aria-label="Main navigation">
          <a href="https://github.com/Grizak/Wick" rel="noreferrer" target="_blank">
            GitHub <span aria-hidden="true">↗</span>
          </a>
          <a className="nav-cta" href="#downloads">
            Get Wick <span aria-hidden="true">↗</span>
          </a>
        </nav>
      </header>

      <main>
        <section className="hero">
          <div className="hero-copy">
            <div className="eyebrow"><span className="live-dot" /> THE WICK CHANGELOG</div>
            <h1>Small language.<br /><span>Big ideas.</span></h1>
            <p>
              A running log of what’s new in Wick — the language that gets out
              of your way and lets you build.
            </p>
            <a className="hero-link" href="#updates">Explore the updates <span aria-hidden="true">↓</span></a>
          </div>
          <div className="hero-art" aria-hidden="true">
            <div className="orbit orbit-one" />
            <div className="orbit orbit-two" />
            <div className="hero-glyph">w<span>.</span></div>
            <div className="hero-chip chip-one">WICK / LANG</div>
            <div className="hero-chip chip-two">COMPILE · RUN · REPEAT</div>
            <div className="hero-star">✳</div>
          </div>
          <div className="hero-foot">
            <span>MADE FOR CURIOSITY</span>
            <span>OPEN SOURCE <span className="hero-foot-mark">✳</span> ALWAYS EVOLVING</span>
          </div>
        </section>

        <section className="updates-section" id="updates">
          <div className="section-heading">
            <div>
              <p className="section-kicker">FROM THE WORKBENCH</p>
              <h2>Latest updates<span className="lime-period">.</span></h2>
            </div>
            <label className="search-box">
              <span className="search-icon" aria-hidden="true">⌕</span>
              <input
                aria-label="Search changelogs"
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search updates"
                ref={searchInput}
                type="search"
                value={search}
              />
              <kbd>/</kbd>
            </label>
          </div>

          {loadingChangelogs ? (
            <div className="status-card"><span className="spinner" /> Loading the latest notes…</div>
          ) : changelogError && changelogs.length === 0 ? (
            <div className="status-card error-card">
              <span className="status-symbol">!</span>
              <span><strong>Changelog unavailable</strong><small>{changelogError}</small></span>
            </div>
          ) : filteredChangelogs.length === 0 ? (
            <div className="status-card">
              {changelogs.length ? 'No updates match that search.' : 'No changelog entries have been published yet.'}
            </div>
          ) : (
            <div className="changelog-list">
              {filteredChangelogs.map((entry, index) => (
                <article className="changelog-entry" key={`${entry.id}-${index}`}>
                  <div className="entry-marker"><span /></div>
                  <div className="entry-main">
                    <div className="entry-meta">
                      {entry.version && <span className="version-pill">v{normalizeVersion(entry.version)}</span>}
                      {formatDate(entry.date) && <time>{formatDate(entry.date)}</time>}
                      {!entry.date && <span className="entry-label">{index === 0 ? 'LATEST NOTES' : 'RELEASE NOTES'}</span>}
                    </div>
                    <h3>{entry.title}</h3>
                    <MarkdownContent content={entry.content} />
                  </div>
                  <span className="entry-index">{String(index + 1).padStart(2, '0')}</span>
                </article>
              ))}
            </div>
          )}
        </section>

        <section className="release-section" id="downloads">
          <div className="release-intro">
            <p className="section-kicker">READY WHEN YOU ARE</p>
            <h2>Pick your<br />Wick<span className="lime-period">.</span></h2>
            <p className="release-description">
              Every release, in one place. Choose a version and grab the right
              build for your machine.
            </p>
            <a
              className="release-github"
              href="https://github.com/Grizak/Wick/releases"
              rel="noreferrer"
              target="_blank"
            >
              Browse on GitHub <span aria-hidden="true">↗</span>
            </a>
            <div className="release-decoration" aria-hidden="true">W</div>
          </div>
          <div className="release-panel">
            <div className="release-panel-header">
              <div>
                <p className="section-kicker">AVAILABLE BUILDS</p>
                <h3>Downloads</h3>
              </div>
              <span className="registry-status"><span className="live-dot" /> LIVE REGISTRY</span>
            </div>
            {loadingRegistry ? (
              <div className="status-card"><span className="spinner" /> Checking the package registry…</div>
            ) : registryError ? (
              <div className="status-card error-card">
                <span className="status-symbol">!</span>
                <span><strong>Couldn’t load releases</strong><small>{registryError}</small></span>
              </div>
            ) : (
              <>
                <div className="version-picker">
                  <label htmlFor="version-select">VERSION</label>
                  <select
                    id="version-select"
                    onChange={(event) => setSelectedVersion(event.target.value)}
                    value={selectedVersion}
                  >
                    {versions.map((release, index) => (
                      <option key={release.version} value={release.version}>
                        {release.version}{index === 0 ? ' — Latest' : ''}
                      </option>
                    ))}
                  </select>
                  <span aria-hidden="true">⌄</span>
                </div>
                {selectedRelease && (
                  <div className="selected-release">
                    <div className="selected-release-title">
                      <span className="release-version">v{selectedRelease.version}</span>
                      <span className="latest-tag">{selectedRelease === versions[0] ? 'LATEST' : 'RELEASE'}</span>
                    </div>
                    <p>{selectedRelease.description ?? 'Wick compiler release'}</p>
                    <PackageDownloads release={selectedRelease} />
                  </div>
                )}
              </>
            )}
            <div className="install-note">
              <span aria-hidden="true">⌘</span>
              <p>Want the source? <a href="https://github.com/Grizak/Wick" rel="noreferrer" target="_blank">Build Wick from GitHub <span aria-hidden="true">↗</span></a></p>
            </div>
          </div>
        </section>
      </main>

      <footer className="footer">
        <a className="brand footer-brand" href="#" aria-label="Wick home">
          <span className="brand-mark" aria-hidden="true">w</span><span>wick<span className="brand-period">.</span></span>
        </a>
        <span>AN OPEN LANGUAGE FOR OPEN MINDS.</span>
        <a href="https://github.com/Grizak/Wick" rel="noreferrer" target="_blank">MADE IN THE OPEN <span aria-hidden="true">↗</span></a>
      </footer>
    </div>
  )
}

export default App
