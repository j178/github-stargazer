import React, {useEffect, useState} from "react";
import styles from './RepoSelector.module.css';

const RepoSelector = ({ repos, onSelect, maxVisible, loadMoreRepos }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [displayedRepos, setDisplayedRepos] = useState([]);

    useEffect(() => {
        setDisplayedRepos(repos.slice(0, maxVisible));
    }, [repos, maxVisible]);

    const handleSearchChange = (e) => {
        setSearchTerm(e.target.value);
        const filteredRepos = repos.filter(repo => repo.toLowerCase().includes(e.target.value.toLowerCase()));
        setDisplayedRepos(filteredRepos.slice(0, maxVisible));
    };

    return (
        <div className={styles.repoSelector}>
            <input
                type="text"
                placeholder="Search Repositories"
                value={searchTerm}
                onChange={handleSearchChange}
                className={styles.searchInput}
            />
            <div className={styles.selectContainer}>
                <select
                    size={10}
                    className={styles.repoSelect}
                    onChange={onSelect}
                >
                    {displayedRepos.map(repo => (
                        <option key={repo} value={repo}>{repo}</option>
                    ))}
                </select>
                <button onClick={loadMoreRepos} className={styles.loadMoreButton}>
                    Load More Repositories
                </button>
            </div>

        </div>
    );
};

export default RepoSelector;
