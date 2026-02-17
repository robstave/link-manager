import { useState } from 'react';
import { api } from '../../services/api';

function normalizeUrl(rawUrl) {
    const value = (rawUrl || '').trim();
    if (!value) return '';
    if (/^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(value)) {
        return value;
    }
    return `https://${value}`;
}

export default function AddLinkModal({ projectId, categories, initialCategoryId, initialCategoryName, onClose, onCreated }) {
    const [url, setUrl] = useState('');
    const [title, setTitle] = useState('');
    const [categoryInput, setCategoryInput] = useState(initialCategoryName || '');
    const [selectedCategoryId, setSelectedCategoryId] = useState(initialCategoryId || '');
    const [description, setDescription] = useState('');
    const [userNotes, setUserNotes] = useState('');
    const [stars, setStars] = useState(0);
    const [error, setError] = useState('');

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

        try {
            await api.createLink({
                url: normalizedUrl,
                title,
                description,
                user_notes: userNotes,
                stars: parseInt(stars) || 0,
                project_id: projectId,
                category_id: catId,
            });
            onCreated();
        } catch (err) {
            setError('Failed to save link: ' + err.message);
        }
    }

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                <h3>Add New Link</h3>
                <form onSubmit={handleSubmit}>
                    <div className="input-group">
                        <label>URL</label>
                        <input
                            type="text"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            required
                            placeholder="www.example.com or https://example.com"
                            autoFocus
                        />
                    </div>
                    <div className="input-group">
                        <label>Title (Optional)</label>
                        <input
                            type="text"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            placeholder="Fetch automatically if empty"
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
                        <label>Description (Optional)</label>
                        <textarea
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            placeholder="Describe why this link matters"
                        />
                    </div>
                    <div className="input-group">
                        <label>User Notes (Optional)</label>
                        <textarea
                            value={userNotes}
                            onChange={(e) => setUserNotes(e.target.value)}
                            placeholder="Personal notes for this link"
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
                        <button type="submit" className="btn btn-primary">Save Link</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
