import { useState } from 'react';
import { api } from '../../services/api';

export default function AddProjectModal({ onClose, onCreated }) {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [error, setError] = useState('');

    async function handleSubmit(e) {
        e.preventDefault();
        setError('');

        try {
            await api.createProject(name, description);
            onCreated();
        } catch (err) {
            setError('Failed to create project: ' + err.message);
        }
    }

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                <h3>Add New Project</h3>
                <form onSubmit={handleSubmit}>
                    <div className="input-group">
                        <label>Project Name</label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                            placeholder="My Awesome Project"
                            autoFocus
                        />
                    </div>
                    <div className="input-group">
                        <label>Description (Optional)</label>
                        <textarea
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            placeholder="What is this project about?"
                        />
                    </div>
                    {error && <p className="modal-error">{error}</p>}
                    <div className="modal-actions">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>Cancel</button>
                        <button type="submit" className="btn btn-primary">Create Project</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
