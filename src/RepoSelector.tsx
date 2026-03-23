import { type FC, useCallback, useEffect, useRef, useState } from 'react'
import { FiArrowRight, FiSearch } from 'react-icons/fi'
import { GoRepo, GoRepoForked } from 'react-icons/go'

import type { RepoInfo } from './models'

import styles from './RepoSelector.module.css'

const RepoSelector: FC<{
  repos: RepoInfo[]
  onSelect: (repoName: string) => void
  loadMoreRepos: () => Promise<void>
  hasMore: boolean
}> = ({ repos, onSelect, loadMoreRepos, hasMore }) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const observer = useRef<IntersectionObserver | null>(null)

  useEffect(() => {
    return () => {
      observer.current?.disconnect()
    }
  }, [])

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

    return (
      repo.name.toLowerCase().includes(normalizedSearch) ||
      (repo.description ?? '').toLowerCase().includes(normalizedSearch)
    )
  })

  return (
    <div className={styles.repoSelector}>
      <div className={styles.searchField}>
        <FiSearch />
        <input
          className={styles.searchInput}
          onChange={(event) => setSearchTerm(event.target.value)}
          placeholder='Search repositories by name or description'
          type='text'
          value={searchTerm}
        />
      </div>

      <div className={styles.repoList}>
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
                <div className={styles.repoCardTop}>
                  <div className={styles.repoIdentity}>
                    <span className={styles.repoIcon}>{repo.fork ? <GoRepoForked /> : <GoRepo />}</span>
                    <strong className={styles.repoName}>{repo.name}</strong>
                  </div>
                  <span className={styles.repoAction}>
                    Add
                    <FiArrowRight />
                  </span>
                </div>
                <p className={styles.repoDescription}>
                  {repo.description?.trim() || 'No repository description provided.'}
                </p>
                <div className={styles.repoMeta}>
                  <span className={styles.repoBadge}>{repo.fork ? 'Fork' : 'Source repository'}</span>
                </div>
              </button>
            )
          })
        )}
      </div>

      <div className={styles.footer}>
        {hasMore ? (
          <button className={styles.loadMoreButton} disabled={isLoading} onClick={() => void loadMore()} type='button'>
            {isLoading ? 'Loading…' : 'Load more repositories'}
          </button>
        ) : (
          <span className={styles.endOfList}>All available repositories are loaded.</span>
        )}
      </div>
    </div>
  )
}

export default RepoSelector
