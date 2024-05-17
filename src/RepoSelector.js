import React, {useEffect, useState} from "react";

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

    const handleLoadMore = async () => {
        await loadMoreRepos();
    };

    return (
        <div>
            <input
                type="text"
                placeholder="Search Repositories"
                value={searchTerm}
                onChange={handleSearchChange}
                style={{ width: "100%", padding: "10px", fontSize: "1em", marginBottom: "10px" }}
            />
            <select
                size={10}
                style={{ width: "100%", padding: "10px", fontSize: "1em", border: "1px solid #ccc", borderRadius: "4px" }}
                onChange={onSelect}
            >
                {displayedRepos.map(repo => (
                    <option key={repo} value={repo}>{repo}</option>
                ))}
            </select>
            <button onClick={handleLoadMore} style={{ width: "100%", padding: "10px", marginTop: "10px" }}>
                Load More Repositories
            </button>
        </div>
    );
};

export default RepoSelector;
