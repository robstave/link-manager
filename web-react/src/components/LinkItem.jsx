import { api } from '../services/api';

export default function LinkItem({ link }) {
    function handleClick() {
        api.recordClick(link.id).catch((err) =>
            console.error('Click tracking failed:', err)
        );
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
                <div className="link-title">{link.title || link.url}</div>
                <div className="link-meta">
                    {starHtml}
                    <span>{link.click_count} clicks</span>
                    {link.tags && link.tags.length > 0 && (
                        <span className="link-tags">
                            {link.tags.slice(0, 2).join(', ')}
                        </span>
                    )}
                </div>
            </div>
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
