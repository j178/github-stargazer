import React, { useCallback, useEffect, useRef, useState } from 'react'
import { GoRepo, GoRepoForked } from 'react-icons/go'

import { RepoInfo } from './models'

import styles from './RepoSelector.module.css'

const RepoSelector: React.FC<{
  repos: RepoInfo[]
  onSelect: (event: React.MouseEvent<HTML>) => void
  loadMoreRepos: () => Promise<void>
  hasMore: boolean
}> = ({ repos, onSelect, loadMoreRepos, hasMore }) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [displayedRepos, setDisplayedRepos] = useState<RepoInfo[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const observer = useRef<IntersectionObserver>()

  useEffect(() => {
    setDisplayedRepos(repos)
  }, [repos])

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value)
    const filteredRepos = repos.filter((repo) => repo.name.toLowerCase().includes(event.target.value.toLowerCase()))
    setDisplayedRepos(filteredRepos)
  }

  const loadMore = useCallback(async () => {
    if (!hasMore || isLoading) return
    setIsLoading(true)
    await loadMoreRepos()
    setIsLoading(false)
  }, [loadMoreRepos, isLoading, hasMore])

  const lastRepoElementRef = useCallback(
    (node: HTMLLIElement | null) => {
      if (isLoading) return
      if (observer.current) observer.current.disconnect()
      observer.current = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting) {
          loadMore()
        }
      })
      if (node) observer.current.observe(node)
    },
    [isLoading, loadMore]
  )

  return (
    <div className={styles.dropdown}>
      <input
        type='text'
        placeholder='Search Repositories'
        value={searchTerm}
        onChange={handleSearchChange}
        className={styles.searchInput}
      />
      <ul className={styles.dropdownList}>
        {displayedRepos.map((repo, index) => {
          if (displayedRepos.length === index + 1) {
            return (
              <li
                ref={lastRepoElementRef}
                key={repo.name}
                value={repo.name}
                className={styles.dropdownOption}
                onClick={onSelect}
              >
                {repo.name}
              </li>
            )
          } else {
            return (
              <li key={repo.name} value={repo.name} className={styles.dropdownOption} onClick={onSelect}>
                {repo.fork ? <GoRepoForked /> : <GoRepo />}
                {repo.owner}/<strong>{repo.name}</strong>
              </li>
            )
          }
        })}
      </ul>

      {isLoading && <div className={styles.spinner}></div>}
    </div>
  )
}

export default RepoSelector
