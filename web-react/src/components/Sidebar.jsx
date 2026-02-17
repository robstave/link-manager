import { useState } from 'react';
import AddProjectModal from './modals/AddProjectModal';

export default function Sidebar({ projects, currentProjectId, onSelectProject, tags, onProjectCreated, onLogout }) {
    const [showAddProject, setShowAddProject] = useState(false);

    return (
        <aside>
            <div className="sidebar-header">
                <span>🔗</span>
                <span>Link Manager</span>
            </div>
            <div className="sidebar-nav">
                <div className="nav-section">
                    <div className="nav-title">Projects</div>
                    <div className="project-list">
                        {(projects || []).map((p) => (
                            <div
                                key={p.id}
                                className={`nav-item ${p.id === currentProjectId ? 'active' : ''}`}
                                onClick={() => onSelectProject(p.id)}
                            >
                                <span className="icon">📁</span>
                                <span>{p.name}</span>
                                <span className="nav-item-count">{p.link_count}</span>
                            </div>
                        ))}
                    </div>
                    <button
                        className="nav-item add-project-btn"
                        onClick={() => setShowAddProject(true)}
                    >
                        <span>+ New Project</span>
                    </button>
                </div>
                <div className="nav-section">
                    <div className="nav-title">Tags</div>
                    <div className="tag-list">
                        {(tags || []).map((t) => (
                            <span
                                key={t.name}
                                className="tag-pill"
                                style={{ background: t.color || 'var(--bg-card)' }}
                            >
                                {t.name} ({t.link_count})
                            </span>
                        ))}
                    </div>
                </div>
            </div>
            <div className="sidebar-footer">
                <div className="nav-item" onClick={onLogout}>
                    <span className="icon">🚪</span>
                    <span>Logout</span>
                </div>
            </div>

            {showAddProject && (
                <AddProjectModal
                    onClose={() => setShowAddProject(false)}
                    onCreated={() => {
                        setShowAddProject(false);
                        onProjectCreated();
                    }}
                />
            )}
        </aside>
    );
}
