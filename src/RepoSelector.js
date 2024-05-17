import React, {useCallback, useEffect, useRef, useState} from "react";
import styles from './RepoSelector.module.css';

const RepoSelector = ({repos, onSelect, loadMoreRepos}) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [displayedRepos, setDisplayedRepos] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const observer = useRef();

    useEffect(() => {
        setDisplayedRepos(repos);
    }, [repos]);

    const handleSearchChange = (e) => {
        setSearchTerm(e.target.value);
        const filteredRepos = repos.filter(repo => repo.toLowerCase().includes(e.target.value.toLowerCase()));
        setDisplayedRepos(filteredRepos);
    };

    const loadMore = useCallback(async () => {
        setIsLoading(true);
        await loadMoreRepos();
        setIsLoading(false);
    }, [loadMoreRepos]);

    const lastRepoElementRef = useCallback(node => {
        if (isLoading) return;
        if (observer.current) observer.current.disconnect();
        observer.current = new IntersectionObserver(entries => {
            if (entries[0].isIntersecting) {
                loadMore();
            }
        });
        if (node) observer.current.observe(node);
    }, [isLoading, loadMore]);

    return (
        <div className={styles.repoSelector}>
            <input type="text"
                   placeholder="Search Repositories"
                   value={searchTerm}
                   onChange={handleSearchChange}
                   className={styles.searchInput}
            />
            <div className={styles.selectContainer}>
                <select size={10}
                        className={styles.repoSelect}
                        onChange={onSelect}
                >
                    {displayedRepos.map((repo, index) => {
                        if (displayedRepos.length === index + 1) {
                            return (
                                <option ref={lastRepoElementRef} key={repo} value={repo}>
                                    {repo}
                                </option>
                            );
                        } else {
                            return (
                                <option key={repo} value={repo}>
                                    {repo}
                                </option>
                            );
                        }
                    })}
                </select>
                {isLoading && <div className={styles.loading}>Loading...</div>}
            </div>
        </div>
    );
};

export default RepoSelector;
