import { useState, useEffect } from 'react';
import { api } from '../services/api';
import LinkItem from './LinkItem';
import AddLinkModal from './modals/AddLinkModal';

export default function CategoryCard({ category, projectId, onLinkCreated, onEdit, refreshTimestamp }) {
    const [links, setLinks] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showAddLink, setShowAddLink] = useState(false);

    useEffect(() => {
        async function fetchLinks() {
            setLoading(true);
            try {
                const data = await api.getLinks({
                    category_id: category.id,
                    limit: 15,
                    sort: 'stars',
                });
                setLinks(data?.links || []);
            } catch (err) {
                console.error('Failed to load links:', err);
            }
            setLoading(false);
        }
        fetchLinks();
    }, [category.id, refreshTimestamp]);

    return (
        <>
            <div className="category-card">
                <div className="category-header">
                    <div className="category-title">
                        <span>{category.is_default ? '📦' : '📁'}</span>
                        <span>{category.name}</span>
                    </div>
                    <div className="category-actions">
                        <button
                            className="btn category-add-link"
                            onClick={() => setShowAddLink(true)}
                        >
                            + Add Link
                        </button>
                        <span className="category-count">{category.link_count}</span>
                    </div>
                </div>
                <div className="link-list">
                    {loading ? (
                        <div className="loading-text">Loading...</div>
                    ) : links.length > 0 ? (
                        links.map((link) => <LinkItem key={link.id} link={link} onEdit={onEdit} />)
                    ) : (
                        <div className="empty-text">No links found</div>
                    )}
                </div>
                {category.link_count > 15 && (
                    <button className="btn view-all-btn">
                        View All {category.link_count} Links
                    </button>
                )}
            </div>

            {showAddLink && (
                <AddLinkModal
                    projectId={projectId}
                    categories={[category]}
                    initialCategoryId={category.id}
                    initialCategoryName={category.name}
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
