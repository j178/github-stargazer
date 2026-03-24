import { type FC, useCallback, useEffect, useRef, useState } from 'react'
import { FiSearch } from 'react-icons/fi'
import { GoRepo, GoRepoForked } from 'react-icons/go'

import type { RepoInfo } from './models'

import styles from './RepoSelector.module.css'

const RepoSelector: FC<{
  repos: RepoInfo[]
  onSelect: (repoName: string) => void
  loadMoreRepos: () => Promise<void>
  hasMore: boolean
  autoFocus?: boolean
  expanded?: boolean
}> = ({ repos, onSelect, loadMoreRepos, hasMore, autoFocus = false, expanded = false }) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [isLoading, setIsLoading] = useState(false)
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
    if (!hasMore || isLoading) {
      return
    }

    setIsLoading(true)
    try {
      await loadMoreRepos()
    } finally {
      setIsLoading(false)
    }
  }, [hasMore, isLoading, loadMoreRepos])

  const lastRepoRef = useCallback(
    (node: HTMLButtonElement | null) => {
      if (isLoading || !hasMore) {
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
    [hasMore, isLoading, loadMore]
  )

  const normalizedSearch = searchTerm.trim().toLowerCase()
  const filteredRepos = repos.filter((repo) => {
    if (!normalizedSearch) {
      return true
    }

    return repo.name.toLowerCase().includes(normalizedSearch)
  })

  useEffect(() => {
    if (normalizedSearch || filteredRepos.length > 0 || !hasMore || isLoading) {
      return
    }

    void loadMore()
  }, [filteredRepos.length, hasMore, isLoading, loadMore, normalizedSearch])

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
        {filteredRepos.length === 0 ? (
          <div className={styles.emptyState}>
            No repositories match the current search.
            {hasMore ? ' Load more results or adjust the search terms.' : ''}
          </div>
        ) : (
          filteredRepos.map((repo, index) => {
            const isLastItem = index === filteredRepos.length - 1
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
