import { api } from '../services/api';

// Truncation limits - adjust these to tweak display
const TITLE_MAX_LENGTH = 20;
const DESCRIPTION_MAX_LENGTH = 20;
const URL_MAX_LENGTH = 25;

export default function LinkItem({ link, onEdit }) {
    function handleClick(e) {
        // Only record click if we're not clicking the edit button
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

    function truncate(text, maxLength) {
        if (!text || text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }

    function cleanUrl(url) {
        if (!url) return '';
        return url.replace(/^https?:\/\//, '').replace(/^www\./, '');
    }

    const starHtml = link.stars > 0 ? (
        <div className="stars">
            <span className="star-text">{link.stars}</span>★
        </div>
    ) : null;

    return (
        <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            className="link-item"
            onClick={handleClick}
        >
            <div className="link-favicon">
                {link.title ? link.title[0].toUpperCase() : '?'}
            </div>
            <div className="link-info">
                <div className="link-title">
                    <span className="link-title-text">{truncate(link.title || link.url, TITLE_MAX_LENGTH)}</span>
                    {link.description && (
                        <span className="link-description-text">
                            {' • '}{truncate(link.description, DESCRIPTION_MAX_LENGTH)}
                        </span>
                    )}
                </div>
                <div className="link-url-line">{truncate(cleanUrl(link.url), URL_MAX_LENGTH)}</div>
                <div className="link-meta">
                    {starHtml}
                    {link.tags && link.tags.length > 0 && (
                        <span className="link-tags">
                            {link.tags.slice(0, 2).join(', ')}
                        </span>
                    )}
                </div>
            </div>
            <button
                className="btn-icon link-edit-btn"
                onClick={handleEditClick}
                title="Edit Link"
            >
                ✏️
            </button>
            <div className="link-hover-info">
                <div className="hover-title">{link.title || link.url}</div>
                <div className="hover-url">{link.url}</div>
                <p className="hover-desc">{link.description || 'No description provided.'}</p>
                <div className="hover-tags">
                    {(link.tags || []).map((t) => (
                        <span key={t} className="hover-tag">{t}</span>
                    ))}
                </div>
            </div>
        </a>
    );
}
