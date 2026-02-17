import { useState, useRef, useEffect } from 'react';
import AddCategoryModal from './modals/AddCategoryModal';
import AddLinkModal from './modals/AddLinkModal';

export default function Header({ user, searchQuery, onSearch, onSearchExecute, currentProjectId, categories, onCategoryCreated, onLinkCreated }) {
    const [showAddCategory, setShowAddCategory] = useState(false);
    const [showAddLink, setShowAddLink] = useState(false);
    const searchTimeout = useRef(null);

    function handleSearchInput(e) {
        const query = e.target.value;
        onSearch(query);

        clearTimeout(searchTimeout.current);
        searchTimeout.current = setTimeout(() => {
            onSearchExecute(query);
        }, 300);
    }

    useEffect(() => {
        return () => clearTimeout(searchTimeout.current);
    }, []);

    return (
        <>
            <header>
                <div className="search-container">
                    <span className="search-icon">🔍</span>
                    <input
                        type="text"
                        placeholder="Search across all links..."
                        value={searchQuery}
                        onChange={handleSearchInput}
                    />
                </div>
                <div className="header-actions">
                    <button className="btn btn-primary" onClick={() => setShowAddCategory(true)}>
                        <span>+ Add Category</span>
                    </button>
                    <button className="btn btn-primary" onClick={() => setShowAddLink(true)}>
                        <span>+ Add Link</span>
                    </button>
                    <div className="user-info">
                        <span className="user-name">{user?.username || 'Admin'}</span>
                        <div className="user-avatar">
                            {(user?.username || 'A')[0].toUpperCase()}
                        </div>
                    </div>
                </div>
            </header>

            {showAddCategory && (
                <AddCategoryModal
                    projectId={currentProjectId}
                    onClose={() => setShowAddCategory(false)}
                    onCreated={() => {
                        setShowAddCategory(false);
                        onCategoryCreated();
                    }}
                />
            )}

            {showAddLink && (
                <AddLinkModal
                    projectId={currentProjectId}
                    categories={categories}
                    onClose={() => setShowAddLink(false)}
                    onCreated={() => {
                        setShowAddLink(false);
                        onLinkCreated();
                    }}
                />
            )}
        </>
    );
}
