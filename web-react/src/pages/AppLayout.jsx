import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../context/AuthContext';
import { api } from '../services/api';
import Sidebar from '../components/Sidebar';
import Header from '../components/Header';
import CategoryGrid from '../components/CategoryGrid';
import SearchResults from '../components/SearchResults';

export default function AppLayout() {
    const { user, logout } = useAuth();
    const [projects, setProjects] = useState([]);
    const [currentProjectId, setCurrentProjectId] = useState(null);
    const [tags, setTags] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState(null);
    const [categories, setCategories] = useState([]);

    const loadProjects = useCallback(async () => {
        const data = await api.getProjects();
        setProjects(data || []);
        return data || [];
    }, []);

    const loadTags = useCallback(async () => {
        const data = await api.getTags();
        setTags(data || []);
    }, []);

    const loadCategories = useCallback(async (projectId) => {
        const data = await api.getCategories(projectId);
        setCategories(data || []);
    }, []);

    // Initial load
    useEffect(() => {
        async function init() {
            const projs = await loadProjects();
            await loadTags();
            if (projs.length > 0) {
                setCurrentProjectId(projs[0].id);
            }
        }
        init();
    }, [loadProjects, loadTags]);

    // Load categories when project changes
    useEffect(() => {
        if (currentProjectId) {
            loadCategories(currentProjectId);
        }
    }, [currentProjectId, loadCategories]);

    const currentProject = projects.find((p) => p.id === currentProjectId);

    async function handleProjectCreated() {
        const projs = await loadProjects();
        // select the newly added project (last one)
        if (projs.length > 0) {
            setCurrentProjectId(projs[projs.length - 1].id);
        }
    }

    async function handleCategoryCreated() {
        await loadCategories(currentProjectId);
    }

    async function handleLinkCreated() {
        await loadCategories(currentProjectId);
        await loadProjects(); // update link counts
    }

    function handleSearch(query) {
        setSearchQuery(query);
        if (!query || query.length < 3) {
            setSearchResults(null);
        }
    }

    async function handleSearchExecute(query) {
        if (query.length > 2) {
            const results = await api.getLinks({ q: query });
            setSearchResults(results?.links || []);
        }
    }

    function handleClearSearch() {
        setSearchQuery('');
        setSearchResults(null);
    }

    return (
        <div className="app-layout">
            <Sidebar
                projects={projects}
                currentProjectId={currentProjectId}
                onSelectProject={(id) => {
                    setCurrentProjectId(id);
                    handleClearSearch();
                }}
                tags={tags}
                onProjectCreated={handleProjectCreated}
                onLogout={logout}
            />
            <main>
                <Header
                    user={user}
                    searchQuery={searchQuery}
                    onSearch={handleSearch}
                    onSearchExecute={handleSearchExecute}
                    currentProjectId={currentProjectId}
                    categories={categories}
                    onCategoryCreated={handleCategoryCreated}
                    onLinkCreated={handleLinkCreated}
                />
                <div className="content-area">
                    {searchResults ? (
                        <SearchResults links={searchResults} />
                    ) : (
                        <>
                            {currentProject && (
                                <div className="project-header">
                                    <h2>{currentProject.name}</h2>
                                    <p>{currentProject.description || 'No description'}</p>
                                </div>
                            )}
                            <CategoryGrid
                                categories={categories}
                                currentProjectId={currentProjectId}
                                onLinkCreated={handleLinkCreated}
                            />
                        </>
                    )}
                </div>
            </main>
        </div>
    );
}
