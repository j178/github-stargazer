import React, { useCallback, useEffect, useRef, useState } from 'react'

import { RepoInfo } from './models'

import styles from './RepoSelector.module.css'

const RepoSelector: React.FC<{
  repos: RepoInfo[]
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void
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
    (node: HTMLOptionElement | null) => {
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
    <div className={styles.repoSelector}>
      <input
        type='text'
        placeholder='Search Repositories'
        value={searchTerm}
        onChange={handleSearchChange}
        className={styles.searchInput}
      />
      <div className={styles.selectContainer}>
        <select size={10} className={styles.repoSelect} onChange={onSelect}>
          <option className={styles.searchOption}>
            <input
              type='text'
              placeholder='Search Repositories'
              value={searchTerm}
              onChange={handleSearchChange}
              className={styles.searchInput}
            />
          </option>
          {displayedRepos.map((repo, index) => {
            if (displayedRepos.length === index + 1) {
              return (
                <option ref={lastRepoElementRef} key={repo.name} value={repo.name}>
                  {repo.name}
                </option>
              )
            } else {
              return (
                <option key={repo.name} value={repo.name}>
                  {repo.name}
                </option>
              )
            }
          })}
        </select>
        {isLoading && <div className={styles.spinner}></div>}
      </div>
    </div>
  )
}

export default RepoSelector
