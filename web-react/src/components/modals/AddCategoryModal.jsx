import { useState } from 'react';
import { api } from '../../services/api';

export default function AddCategoryModal({ projectId, onClose, onCreated }) {
    const [name, setName] = useState('');
    const [error, setError] = useState('');

    async function handleSubmit(e) {
        e.preventDefault();
        setError('');

        if (!projectId) {
            setError('Please select a project first.');
            return;
        }

        try {
            await api.createCategory(projectId, name.trim());
            onCreated();
        } catch (err) {
            setError('Failed to create category: ' + err.message);
        }
    }

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                <h3>Add New Category</h3>
                <form onSubmit={handleSubmit}>
                    <div className="input-group">
                        <label>Category Name</label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                            placeholder="Backend, Docs, Design..."
                            autoFocus
                        />
                    </div>
                    {error && <p className="modal-error">{error}</p>}
                    <div className="modal-actions">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>Cancel</button>
                        <button type="submit" className="btn btn-primary">Create Category</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
