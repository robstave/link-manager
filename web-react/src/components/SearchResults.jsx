import LinkItem from './LinkItem';

export default function SearchResults({ links, refreshTimestamp }) {
    return (
        <div className="search-results">
            <div className="project-header">
                <h2>Search Results</h2>
                <p>Found {links.length} matches</p>
            </div>
            <div className="category-card search-results-card">
                <div className="link-list">
                    {links.length > 0 ? (
                        links.map((link) => <LinkItem key={link.id} link={link} />)
                    ) : (
                        <div className="empty-text">No matches found</div>
                    )}
                </div>
            </div>
        </div>
    );
}
