import { useState, useEffect } from 'react';
import { api } from '../../services/api';

function normalizeUrl(rawUrl) {
    const value = (rawUrl || '').trim();
    if (!value) return '';
    if (/^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(value)) {
        return value;
    }
    return `https://${value}`;
}

export default function EditLinkModal({ link, projects, categories, onClose, onUpdated }) {
    const [url, setUrl] = useState(link.url || '');
    const [title, setTitle] = useState(link.title || '');
    const [categoryInput, setCategoryInput] = useState(link.category?.name || '');
    const [selectedCategoryId, setSelectedCategoryId] = useState(link.category_id || '');
    const [description, setDescription] = useState(link.description || '');
    const [userNotes, setUserNotes] = useState(link.user_notes || '');
    const [stars, setStars] = useState(link.stars || 0);
    const [tagsInput, setTagsInput] = useState((link.tags || []).join(', '));
    const [error, setError] = useState('');
    const [loadingTitle, setLoadingTitle] = useState(false);

    async function handleUrlBlur() {
        if (!url.trim() || title.trim()) return;
        const normalizedUrl = normalizeUrl(url);
        if (!normalizedUrl) return;

        setLoadingTitle(true);
        try {
            const result = await api.fetchTitle(normalizedUrl);
            if (result && result.title && !title.trim()) {
                setTitle(result.title);
            }
        } catch (err) {
            console.error('Failed to fetch title:', err);
        } finally {
            setLoadingTitle(false);
        }
    }

    function handleCategoryInput(e) {
        const value = e.target.value;
        setCategoryInput(value);
        const matched = (categories || []).find((c) => c.name === value);
        setSelectedCategoryId(matched ? matched.id : '');
    }

    async function handleSubmit(e) {
        e.preventDefault();
        setError('');

        const normalizedUrl = normalizeUrl(url);
        const catId = selectedCategoryId || (categories || []).find((c) => c.name === categoryInput.trim())?.id || null;
        const tags = tagsInput.split(',').map(t => t.trim()).filter(t => t !== '');

        try {
            await api.updateLink(link.id, {
                url: normalizedUrl,
                title,
                description,
                user_notes: userNotes,
                stars: parseInt(stars) || 0,
                project_id: link.project_id,
                category_id: catId,
                tags: tags
            });
            onUpdated();
        } catch (err) {
            setError('Failed to update link: ' + err.message);
        }
    }

    async function handleDelete() {
        if (!window.confirm('Are you sure you want to delete this link?')) return;
        try {
            await api.deleteLink(link.id);
            onUpdated();
        } catch (err) {
            setError('Failed to delete link: ' + err.message);
        }
    }

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h3>Edit Link</h3>
                    <button className="btn-icon" onClick={handleDelete} title="Delete Link" style={{ color: '#ef4444' }}>
                        🗑️
                    </button>
                </div>
                <form onSubmit={handleSubmit}>
                    <div className="input-group">
                        <label>URL</label>
                        <input
                            type="text"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            onBlur={handleUrlBlur}
                            required
                        />
                    </div>
                    <div className="input-group">
                        <label>Title</label>
                        <input
                            type="text"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            placeholder={loadingTitle ? "Fetching title..." : ""}
                            disabled={loadingTitle}
                        />
                    </div>
                    <div className="input-group">
                        <label>Category</label>
                        <input
                            type="text"
                            value={categoryInput}
                            onChange={handleCategoryInput}
                            list="category-options"
                            placeholder="Type to filter categories"
                            autoComplete="off"
                        />
                        <datalist id="category-options">
                            {(categories || []).map((cat) => (
                                <option key={cat.id} value={cat.name} />
                            ))}
                        </datalist>
                    </div>
                    <div className="input-group">
                        <label>Tags (comma separated)</label>
                        <input
                            type="text"
                            value={tagsInput}
                            onChange={(e) => setTagsInput(e.target.value)}
                            placeholder="dev, docs, research"
                        />
                    </div>
                    <div className="input-group">
                        <label>Description</label>
                        <textarea
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </div>
                    <div className="input-group">
                        <label>User Notes</label>
                        <textarea
                            value={userNotes}
                            onChange={(e) => setUserNotes(e.target.value)}
                        />
                    </div>
                    <div className="input-group">
                        <label>Stars (0-10)</label>
                        <input
                            type="number"
                            value={stars}
                            onChange={(e) => setStars(e.target.value)}
                            min="0"
                            max="10"
                        />
                    </div>
                    {error && <p className="modal-error">{error}</p>}
                    <div className="modal-actions">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>Cancel</button>
                        <button type="submit" className="btn btn-primary">Save Changes</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
