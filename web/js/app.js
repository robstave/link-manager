document.addEventListener('DOMContentLoaded', async () => {
    const loginPage = document.getElementById('login-page');
    const appPage = document.getElementById('app');
    const loginForm = document.getElementById('login-form');
    const loginError = document.getElementById('login-error');

    // State
    let projects = [];
    let currentProjectId = null;
    let categories = [];

    // Initial Auth Check
    if (api.isAuthenticated()) {
        showApp();
    } else {
        showLogin();
    }

    // Login Handler
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = e.target.username.value;
        const password = e.target.password.value;

        try {
            await api.login(username, password);
            showApp();
        } catch (err) {
            loginError.textContent = err.message;
            loginError.classList.remove('hidden');
        }
    });

    document.getElementById('logout-btn').addEventListener('click', () => {
        api.logout();
    });

    async function showLogin() {
        loginPage.classList.remove('hidden');
        appPage.classList.add('hidden');
    }

    async function showApp() {
        loginPage.classList.add('hidden');
        appPage.classList.remove('hidden');

        try {
            const me = await api.getMe();
            document.getElementById('display-username').textContent = me.username;

            await loadProjects();
            await loadTags();
        } catch (err) {
            console.error('Failed to init app:', err);
            api.logout();
        }
    }

    // Project Logic
    async function loadProjects() {
        projects = await api.getProjects();
        renderProjectSidebar();

        if (projects.length > 0 && !currentProjectId) {
            selectProject(projects[0].id);
        }
    }

    function renderProjectSidebar() {
        const list = document.getElementById('project-list');
        list.innerHTML = '';

        (projects || []).forEach(p => {
            const el = document.createElement('div');
            el.className = `nav-item ${p.id === currentProjectId ? 'active' : ''}`;
            el.innerHTML = `
                <span class="icon">📁</span>
                <span>${p.name}</span>
                <span style="margin-left: auto; opacity: 0.5; font-size: 0.7rem;">${p.link_count}</span>
            `;
            el.onclick = () => selectProject(p.id);
            list.appendChild(el);
        });
    }

    async function selectProject(id) {
        currentProjectId = id;
        const project = projects.find(p => p.id === id);

        document.getElementById('current-project-name').textContent = project.name;
        document.getElementById('current-project-desc').textContent = project.description || 'No description';

        renderProjectSidebar();
        await loadProjectContent(id);
    }

    async function loadProjectContent(projectId) {
        const grid = document.getElementById('category-grid');
        grid.innerHTML = '<div style="padding: 2rem; color: var(--text-muted);">Loading links...</div>';

        const projectCategories = await api.getCategories(projectId);
        categories = projectCategories || [];
        grid.innerHTML = '';

        for (const cat of categories) {
            const card = await renderCategoryCard(cat);
            grid.appendChild(card);
        }
    }

    async function renderCategoryCard(category) {
        const linksData = await api.getLinks({
            category_id: category.id,
            limit: 15,
            sort: 'stars'
        });

        const card = document.createElement('div');
        card.className = 'category-card';

        const linksHtml = (linksData && linksData.links && linksData.links.length > 0)
            ? linksData.links.map(l => renderLinkItem(l)).join('')
            : '<div style="padding: 1rem; text-align: center; color: var(--text-muted); font-size: 0.875rem;">No links found</div>';

        card.innerHTML = `
            <div class="category-header">
                <div class="category-title">
                    <span>${category.is_default ? '📦' : '📁'}</span>
                    <span>${category.name}</span>
                </div>
                <div class="category-actions">
                    <button class="btn category-add-link" data-category-id="${category.id}" data-category-name="${category.name}">+ Add Link</button>
                    <span style="font-size: 0.875rem; color: var(--text-muted);">${category.link_count}</span>
                </div>
            </div>
            <div class="link-list">
                ${linksHtml}
            </div>
            ${category.link_count > 15 ? `<button class="btn" style="margin-top: 1rem; width: 100%; font-size: 0.75rem; justify-content: center; background: rgba(255,255,255,0.05);">View All ${category.link_count} Links</button>` : ''}
        `;

        const categoryAddBtn = card.querySelector('.category-add-link');
        if (categoryAddBtn) {
            categoryAddBtn.onclick = () => openAddLinkModal({
                categoryId: category.id,
                categoryName: category.name,
            });
        }

        return card;
    }

    function renderLinkItem(link) {
        const starHtml = link.stars > 0
            ? `<div class="stars"><span class="star-text">${link.stars}</span>★</div>`
            : '';

        return `
            <a href="${link.url}" target="_blank" class="link-item" onclick="handleLinkClick(event, '${link.id}', '${link.url}')">
                <div class="link-favicon">${link.title ? link.title[0].toUpperCase() : '?'}</div>
                <div class="link-info">
                    <div class="link-title">${link.title || link.url}</div>
                    <div class="link-meta">
                        ${starHtml}
                        <span>${link.click_count} clicks</span>
                        ${link.tags ? `<span style="color: var(--primary);">${link.tags.slice(0, 2).join(', ')}</span>` : ''}
                    </div>
                </div>
                <div class="link-hover-info">
                    <div style="font-weight: 600; margin-bottom: 0.5rem;">${link.title || link.url}</div>
                    <div style="font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.5rem; word-break: break-all;">${link.url}</div>
                    <p style="font-size: 0.8125rem; margin-bottom: 0.75rem;">${link.description || 'No description provided.'}</p>
                    <div style="display: flex; gap: 0.5rem; flex-wrap: wrap;">
                        ${(link.tags || []).map(t => `<span style="background: var(--bg-main); padding: 2px 6px; border-radius: 4px; font-size: 0.7rem;">${t}</span>`).join('')}
                    </div>
                </div>
            </a>
        `;
    }

    // Global helper for click tracking
    window.handleLinkClick = async (e, id, url) => {
        // We let the link open normally, but fire off the click increment in the background
        api.recordClick(id).catch(err => console.error('Click tracking failed:', err));
    };

    // Tags
    async function loadTags() {
        const tags = await api.getTags();
        const list = document.getElementById('tag-list');
        list.innerHTML = '';

        (tags || []).forEach(t => {
            const el = document.createElement('span');
            el.style.cssText = `
                padding: 4px 10px;
                background: ${t.color || 'var(--bg-card)'};
                border: 1px solid var(--border);
                border-radius: 12px;
                font-size: 0.75rem;
                cursor: pointer;
            `;
            el.textContent = `${t.name} (${t.link_count})`;
            list.appendChild(el);
        });
    }

    // Add Project Modal
    const addProjectBtn = document.getElementById('add-project-btn');
    const projectModal = document.getElementById('add-project-modal');
    const closeProjectModalBtn = document.getElementById('close-project-modal-btn');
    const addProjectForm = document.getElementById('add-project-form');

    addProjectBtn.onclick = () => projectModal.classList.remove('hidden');
    closeProjectModalBtn.onclick = () => projectModal.classList.add('hidden');

    addProjectForm.onsubmit = async (e) => {
        e.preventDefault();
        const name = document.getElementById('project-name').value;
        const description = document.getElementById('project-description').value;

        try {
            await api.createProject(name, description);
            projectModal.classList.add('hidden');
            addProjectForm.reset();
            await loadProjects();
        } catch (err) {
            alert('Failed to create project: ' + err.message);
        }
    };

    // Add Category Modal
    const addCategoryBtn = document.getElementById('add-category-btn');
    const addCategoryModal = document.getElementById('add-category-modal');
    const closeCategoryModalBtn = document.getElementById('close-category-modal-btn');
    const addCategoryForm = document.getElementById('add-category-form');

    addCategoryBtn.onclick = () => addCategoryModal.classList.remove('hidden');
    closeCategoryModalBtn.onclick = () => addCategoryModal.classList.add('hidden');

    addCategoryForm.onsubmit = async (e) => {
        e.preventDefault();
        const name = document.getElementById('category-name').value.trim();

        try {
            await api.createCategory(currentProjectId, name);
            addCategoryModal.classList.add('hidden');
            addCategoryForm.reset();
            await loadProjectContent(currentProjectId);
        } catch (err) {
            alert('Failed to create category: ' + err.message);
        }
    };

    // Add Link Modal
    const addLinkBtn = document.getElementById('add-link-btn');
    const modal = document.getElementById('add-link-modal');
    const closeModalBtn = document.getElementById('close-modal-btn');
    const addLinkForm = document.getElementById('add-link-form');
    const linkCategoryInput = document.getElementById('link-category');
    const categoryOptions = document.getElementById('category-options');

    function refreshCategoryOptions(filterText = '') {
        categoryOptions.innerHTML = '';
        const filter = filterText.trim().toLowerCase();
        (categories || [])
            .filter(cat => !filter || cat.name.toLowerCase().includes(filter))
            .forEach(cat => {
                const option = document.createElement('option');
                option.value = cat.name;
                option.dataset.categoryId = cat.id;
                categoryOptions.appendChild(option);
            });
    }

    function openAddLinkModal({ categoryId = null, categoryName = '' } = {}) {
        refreshCategoryOptions();
        linkCategoryInput.value = categoryName || '';
        if (categoryId) {
            linkCategoryInput.dataset.selectedCategoryId = categoryId;
        } else {
            delete linkCategoryInput.dataset.selectedCategoryId;
        }
        modal.classList.remove('hidden');
    }

    addLinkBtn.onclick = () => openAddLinkModal();
    closeModalBtn.onclick = () => modal.classList.add('hidden');

    linkCategoryInput.addEventListener('input', (e) => {
        const inputValue = e.target.value;
        const matched = (categories || []).find(cat => cat.name === inputValue);
        if (matched) {
            linkCategoryInput.dataset.selectedCategoryId = matched.id;
        } else {
            delete linkCategoryInput.dataset.selectedCategoryId;
        }
        refreshCategoryOptions(inputValue);
    });

    addLinkForm.onsubmit = async (e) => {
        e.preventDefault();
        const url = document.getElementById('link-url').value;
        const title = document.getElementById('link-title').value;
        const description = document.getElementById('link-description').value;
        const userNotes = document.getElementById('link-user-notes').value;
        const stars = parseInt(document.getElementById('link-stars').value);
        const selectedCategory = (categories || []).find(cat => cat.name === linkCategoryInput.value.trim());
        const categoryId = linkCategoryInput.dataset.selectedCategoryId || selectedCategory?.id || null;

        try {
            await api.createLink({
                url,
                title,
                description,
                user_notes: userNotes,
                stars,
                project_id: currentProjectId,
                category_id: categoryId
            });
            modal.classList.add('hidden');
            addLinkForm.reset();
            delete linkCategoryInput.dataset.selectedCategoryId;
            await loadProjectContent(currentProjectId);
            await loadProjects(); // Update counts
        } catch (err) {
            alert('Failed to save link: ' + err.message);
        }
    };

    // Global Search
    const searchInput = document.getElementById('global-search');
    let searchTimeout;

    searchInput.oninput = (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(async () => {
            const query = e.target.value;
            if (query.length > 2) {
                const results = await api.getLinks({ q: query });
                renderSearchResults(results.links);
            } else if (query.length === 0) {
                selectProject(currentProjectId);
            }
        }, 300);
    };

    function renderSearchResults(links) {
        document.getElementById('current-project-name').textContent = 'Search Results';
        document.getElementById('current-project-desc').textContent = `Found ${links.length} matches`;

        const grid = document.getElementById('category-grid');
        grid.innerHTML = '';

        const card = document.createElement('div');
        card.className = 'category-card';
        card.style.gridColumn = '1 / -1';

        const linksHtml = links.length > 0
            ? links.map(l => renderLinkItem(l)).join('')
            : '<div style="padding: 2rem; text-align: center; color: var(--text-muted);">No matches found</div>';

        card.innerHTML = `
            <div class="link-list">
                ${linksHtml}
            </div>
        `;
        grid.appendChild(card);
    }
});
