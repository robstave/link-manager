import { api } from '../services/api';

function cleanUrl(url) {
    if (!url) return '';
    return url.replace(/^https?:\/\//, '').replace(/^www\./, '');
}

function formatDate(value) {
    if (!value) return 'Never';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return 'Never';
    return date.toLocaleString();
}

export default function LinkDetailCard({ link, onEdit }) {
    function handleClick(e) {
        if (e.target.closest('.link-edit-btn')) return;
        api.recordClick(link.id).catch((err) =>
            console.error('Click tracking failed:', err)
        );
    }

    function handleEditClick(e) {
        e.preventDefault();
        e.stopPropagation();
        if (onEdit) onEdit(link);
    }

    return (
        <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            className="link-detail-card"
            onClick={handleClick}
        >
            <div className="link-detail-header">
                <div className="link-detail-favicon">
                    {link.icon_url ? (
                        <img
                            src={link.icon_url}
                            alt=""
                            onError={(e) => {
                                e.target.style.display = 'none';
                                e.target.parentElement.textContent = link.title ? link.title[0].toUpperCase() : '?';
                            }}
                        />
                    ) : (
                        link.title ? link.title[0].toUpperCase() : '?'
                    )}
                </div>
                <button
                    className="btn-icon link-edit-btn"
                    onClick={handleEditClick}
                    title="Edit Link"
                >
                    ✏️
                </button>
            </div>

            <div className="link-detail-title">{link.title || link.url}</div>
            <div className="link-detail-url">{cleanUrl(link.url)}</div>

            <p className="link-detail-description">{link.description || 'No description provided.'}</p>

            <p className="link-detail-notes">
                <strong>Notes:</strong> {link.user_notes || 'No user notes.'}
            </p>

            {(link.tags || []).length > 0 && (
                <div className="link-detail-tags">
                    {link.tags.map((tag) => (
                        <span key={tag} className="link-detail-tag">{tag}</span>
                    ))}
                </div>
            )}

            <div className="link-detail-footer">
                <span>⭐ {link.stars}</span>
                <span>👆 {link.click_count} clicks</span>
                <span>🕒 {formatDate(link.last_clicked_at)}</span>
            </div>
        </a>
    );
}
