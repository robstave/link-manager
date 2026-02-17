import { useEffect, useState } from 'react';
import { api } from '../services/api';
import LinkDetailCard from './LinkDetailCard';

export default function CategoryDetailView({ category, projectName, onBack, onEdit, refreshTimestamp }) {
    const [links, setLinks] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchLinks() {
            setLoading(true);
            try {
                const data = await api.getLinks({
                    category_id: category.id,
                    limit: 200,
                    sort: 'stars',
                });
                setLinks(data?.links || []);
            } catch (err) {
                console.error('Failed to load category detail links:', err);
            }
            setLoading(false);
        }

        fetchLinks();
    }, [category.id, refreshTimestamp]);

    return (
        <div>
            <div className="category-detail-header">
                <button className="btn btn-ghost" onClick={onBack}>← Back to Categories</button>
                <div>
                    <h2>{category.name}</h2>
                    <p>{projectName} · {links.length} links</p>
                </div>
            </div>

            {loading ? (
                <div className="loading-text">Loading category links...</div>
            ) : links.length === 0 ? (
                <div className="empty-state">No links in this category yet.</div>
            ) : (
                <div className="category-detail-grid">
                    {links.map((link) => (
                        <LinkDetailCard key={link.id} link={link} onEdit={onEdit} />
                    ))}
                </div>
            )}
        </div>
    );
}
