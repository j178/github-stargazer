import { type FC, useCallback, useEffect, useRef, useState } from 'react'
import { FiSearch } from 'react-icons/fi'
import { GoRepo, GoRepoForked } from 'react-icons/go'

import type { RepoInfo } from './models'

import styles from './RepoSelector.module.css'

const RepoSelector: FC<{
  repos: RepoInfo[]
  onSelect: (repoName: string) => void
  loadMoreRepos: () => Promise<void>
  searchRepos: (query: string) => Promise<RepoInfo[]>
  hasMore: boolean
  isLoading: boolean
  autoFocus?: boolean
  expanded?: boolean
}> = ({ repos, onSelect, loadMoreRepos, searchRepos, hasMore, isLoading, autoFocus = false, expanded = false }) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<RepoInfo[]>([])
  const [searchError, setSearchError] = useState('')
  const observer = useRef<IntersectionObserver | null>(null)
  const inputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    return () => {
      observer.current?.disconnect()
    }
  }, [])

  useEffect(() => {
    if (!autoFocus) {
      return
    }

    const focusTimer = window.setTimeout(() => {
      inputRef.current?.focus()
    }, 0)

    return () => {
      window.clearTimeout(focusTimer)
    }
  }, [autoFocus])

  const loadMore = useCallback(async () => {
    if (!hasMore || isLoadingMore) {
      return
    }

    setIsLoadingMore(true)
    try {
      await loadMoreRepos()
    } finally {
      setIsLoadingMore(false)
    }
  }, [hasMore, isLoadingMore, loadMoreRepos])

  const normalizedSearch = searchTerm.trim().toLowerCase()
  const displayedRepos = normalizedSearch ? searchResults : repos

  const lastRepoRef = useCallback(
    (node: HTMLButtonElement | null) => {
      if (isLoading || isLoadingMore || !hasMore || normalizedSearch) {
        return
      }

      observer.current?.disconnect()
      observer.current = new IntersectionObserver((entries) => {
        if (entries[0]?.isIntersecting) {
          void loadMore()
        }
      })

      if (node) {
        observer.current.observe(node)
      }
    },
    [hasMore, isLoading, isLoadingMore, loadMore, normalizedSearch]
  )

  useEffect(() => {
    if (normalizedSearch || repos.length > 0 || !hasMore || isLoading || isLoadingMore) {
      return
    }

    void loadMore()
  }, [repos.length, hasMore, isLoading, isLoadingMore, loadMore, normalizedSearch])

  useEffect(() => {
    if (!normalizedSearch) {
      setSearchResults([])
      setSearchError('')
      setIsSearching(false)
      return
    }

    let isActive = true
    const searchTimer = window.setTimeout(async () => {
      setIsSearching(true)
      setSearchError('')
      try {
        const results = await searchRepos(normalizedSearch)
        if (!isActive) {
          return
        }
        setSearchResults(results)
      } catch {
        if (!isActive) {
          return
        }
        setSearchResults([])
        setSearchError('Failed to search repositories.')
      } finally {
        if (isActive) {
          setIsSearching(false)
        }
      }
    }, 220)

    return () => {
      isActive = false
      window.clearTimeout(searchTimer)
    }
  }, [normalizedSearch, searchRepos])

  return (
    <div className={styles.repoSelector}>
      <div className={styles.searchField}>
        <FiSearch />
        <input
          className={styles.searchInput}
          onChange={(event) => setSearchTerm(event.target.value)}
          placeholder='Search repositories by name'
          ref={inputRef}
          type='text'
          value={searchTerm}
        />
      </div>

      <div className={expanded ? `${styles.repoList} ${styles.repoListExpanded}` : styles.repoList}>
        {isSearching && displayedRepos.length === 0 ? (
          <div className={styles.emptyState}>Searching repositories…</div>
        ) : isLoading && displayedRepos.length === 0 && !normalizedSearch ? (
          null
        ) : searchError ? (
          <div className={styles.emptyState}>{searchError}</div>
        ) : displayedRepos.length === 0 ? (
          <div className={styles.emptyState}>
            {normalizedSearch
              ? 'No repositories match the current search.'
              : 'No more repositories are available to add.'}
          </div>
        ) : (
          displayedRepos.map((repo, index) => {
            const isLastItem = index === displayedRepos.length - 1
            return (
              <button
                className={styles.repoCard}
                key={repo.name}
                onClick={() => onSelect(repo.name)}
                ref={isLastItem ? lastRepoRef : undefined}
                type='button'
              >
                <div className={styles.repoIdentity}>
                  <span className={styles.repoIcon}>{repo.fork ? <GoRepoForked /> : <GoRepo />}</span>
                  <strong className={styles.repoName}>{repo.name}</strong>
                </div>
              </button>
            )
          })
        )}
      </div>
    </div>
  )
}

export default RepoSelector
