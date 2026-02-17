import CategoryCard from './CategoryCard';

export default function CategoryGrid({ categories, currentProjectId, onLinkCreated, refreshTimestamp }) {
    if (!categories || categories.length === 0) {
        return (
            <div className="empty-state">
                No categories yet — click "+ Add Category" to get started.
            </div>
        );
    }

    return (
        <div className="category-grid">
            {categories.map((cat) => (
                <CategoryCard
                    key={cat.id}
                    category={cat}
                    projectId={currentProjectId}
                    onLinkCreated={onLinkCreated}
                    refreshTimestamp={refreshTimestamp}
                />
            ))}
        </div>
    );
}
