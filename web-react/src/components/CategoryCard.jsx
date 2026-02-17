import { useState, useEffect, useRef } from 'react';
import { api } from '../services/api';
import LinkItem from './LinkItem';
import AddLinkModal from './modals/AddLinkModal';

function extractDroppedUrl(e) {
    // Try text/uri-list first (standard browser drag from address bar)
    let url = e.dataTransfer.getData('text/uri-list');
    if (url) {
        // uri-list can have multiple lines; skip comments (#)
        const lines = url.split('\n').filter(l => l.trim() && !l.startsWith('#'));
        if (lines.length > 0) return lines[0].trim();
    }
    // Fall back to text/plain
    url = e.dataTransfer.getData('text/plain');
    if (url) {
        url = url.trim();
        // Basic validation: must look like a URL
        if (url.startsWith('http://') || url.startsWith('https://') || url.includes('.')) {
            return url;
        }
    }
    return null;
}

export default function CategoryCard({ category, projectId, onLinkCreated, onEdit, onOpenCategory, refreshTimestamp }) {
    const [links, setLinks] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showAddLink, setShowAddLink] = useState(false);
    const [dragOver, setDragOver] = useState(false);
    const dragCounter = useRef(0);

    function handleDragEnter(e) {
        e.preventDefault();
        dragCounter.current++;
        setDragOver(true);
    }

    function handleDragLeave(e) {
        e.preventDefault();
        dragCounter.current--;
        if (dragCounter.current <= 0) {
            dragCounter.current = 0;
            setDragOver(false);
        }
    }

    function handleDragOver(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
    }

    async function handleDrop(e) {
        e.preventDefault();
        dragCounter.current = 0;
        setDragOver(false);

        const url = extractDroppedUrl(e);
        if (!url) return;

        try {
            await api.createLink({
                url,
                project_id: projectId,
                category_id: category.id,
            });
            if (onLinkCreated) onLinkCreated();
        } catch (err) {
            console.error('Failed to create link from drop:', err);
        }
    }

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
            <div
                className={`category-card${dragOver ? ' drag-over' : ''}`}
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onDragOver={handleDragOver}
                onDrop={handleDrop}
            >
                <div className="category-header">
                    <button
                        className="category-title category-title-link"
                        type="button"
                        onClick={() => onOpenCategory?.(category)}
                    >
                        <span>{category.is_default ? '📦' : '📁'}</span>
                        <span>{category.name}</span>
                    </button>
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
                    <button className="btn view-all-btn" onClick={() => onOpenCategory?.(category)}>
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
