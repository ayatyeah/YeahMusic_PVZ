const API_URL = '/api';
const audio = document.getElementById('audio');

let state = {
    user: JSON.parse(localStorage.getItem('user')),
    albums: [],
    singles: [],
    playlists: [],
    queue: [],
    currentIndex: 0,
    isPlaying: false,
    lyricsData: [],
    isKaraoke: false,
    currentTrackId: null,
    lastOpenedTrack: null
};

let tapState = {
    on: false,
    lines: [],
    times: [],
    idx: 0
};

document.addEventListener('DOMContentLoaded', () => {
    lucide.createIcons();
    init();
    window.addEventListener('popstate', (e) => routeFromState(e.state));
});

function init() {
    setupUI();
    setupAudio();
    fetchContent().then(() => {
        if (!history.state) history.replaceState({ view: 'home' }, '', '#home');
        routeFromState(history.state);
    });
}

function setActiveNav(key) {
    document.querySelectorAll('.bottom-nav .nav-item').forEach(b => {
        if (b.dataset.nav === key) b.classList.add('active');
        else b.classList.remove('active');
    });
}

function setupUI() {
    if (state.user) {
        const authBtn = document.getElementById('auth-btn');
        const userAvatar = document.getElementById('user-avatar');
        if (authBtn) authBtn.classList.add('hidden');
        if (userAvatar) userAvatar.classList.remove('hidden');

        const letter = (state.user.name || 'U')[0].toUpperCase();

        const ua = document.getElementById('user-avatar');
        const pa = document.getElementById('p-avatar');
        const pn = document.getElementById('p-name');
        const pe = document.getElementById('p-email');

        if (ua) ua.innerText = letter;
        if (pa) pa.innerText = letter;
        if (pn) pn.innerText = state.user.name || 'User';
        if (pe) pe.innerText = state.user.email || '';
    }
}

async function fetchContent() {
    try {
        const res = await fetch(`${API_URL}/content`);
        const data = await res.json();
        state.albums = data.albums || [];
        state.singles = data.singles || [];
        state.playlists = data.playlists || [];
    } catch (e) {
        console.error(e);
        state.albums = [];
        state.singles = [];
        state.playlists = [];
    }
}

function navigateHome() {
    history.pushState({ view: 'home' }, '', '#home');
    routeFromState({ view: 'home' });
}

function navigateSearch() {
    history.pushState({ view: 'search' }, '', '#search');
    routeFromState({ view: 'search' });
}

function navigateLibrary() {
    history.pushState({ view: 'library' }, '', '#library');
    routeFromState({ view: 'library' });
}

function navigateSuno() {
    history.pushState({ view: 'suno' }, '', '#suno');
    routeFromState({ view: 'suno' });
}

function navigateEditProfile() {
    closeModal('profile-modal');
    history.pushState({ view: 'edit-profile' }, '', '#edit-profile');
    routeFromState({ view: 'edit-profile' });
}

function navigateCreateAlbum() {
    history.pushState({ view: 'create-album' }, '', '#create-album');
    routeFromState({ view: 'create-album' });
}

function navigateUploadTrack() {
    history.pushState({ view: 'upload-track' }, '', '#upload-track');
    routeFromState({ view: 'upload-track' });
}

function navigateOpenPage(type, id) {
    history.pushState({ view: 'page', type, id }, '', `#${type}/${id}`);
    routeFromState({ view: 'page', type, id });
}

function navigateEditLyrics(trackId, existingLyrics, title, artist) {
    history.pushState({ view: 'edit-lyrics', trackId, existingLyrics, title, artist }, '', `#edit-lyrics/${trackId}`);
    routeFromState({ view: 'edit-lyrics', trackId, existingLyrics, title, artist });
}

function goBack() {
    history.back();
}

function routeFromState(s) {
    if (!s || !s.view) {
        renderHome();
        return;
    }
    if (s.view === 'home') renderHome();
    else if (s.view === 'search') renderSearchPage();
    else if (s.view === 'library') renderLibrary();
    else if (s.view === 'suno') renderSunoToolsPage();
    else if (s.view === 'edit-profile') renderEditProfile();
    else if (s.view === 'create-album') renderCreateAlbum();
    else if (s.view === 'upload-track') renderUploadTrack();
    else if (s.view === 'page') openPage(s.type, s.id, true);
    else if (s.view === 'edit-lyrics') renderEditLyrics(s);
    else renderHome();
}

function renderHome() {
    setActiveNav('home');
    const main = document.getElementById('main-view');

    const time = new Date().getHours();
    let greeting = "Good Morning";
    if (time >= 12) greeting = "Good Afternoon";
    if (time >= 18) greeting = "Good Evening";

    main.innerHTML = `
        <div class="hero-banner">
            <h1>${greeting}, ${state.user ? (escapeHtml(state.user.name || 'User')) : 'Guest'}</h1>
            <p>Discover new vibes today.</p>
        </div>

        ${state.albums.length === 0 && state.singles.length === 0 && state.playlists.length === 0 ? '<p style="opacity:0.65;font-weight:950;margin-top:16px;">Nothing here yet.</p>' : ''}

        <span class="section-h">New Albums</span>
        <div class="grid-cards">${state.albums.map(a => card(a, 'album')).join('')}</div>

        <span class="section-h">Hot Singles</span>
        <div class="grid-cards">${state.singles.map(s => card(s, 'album')).join('')}</div>

        <span class="section-h">Playlists</span>
        <div class="grid-cards">${state.playlists.map(p => card(p, 'playlist')).join('')}</div>
    `;
    closeAll();
    lucide.createIcons();
}

function card(item, type) {
    return `
        <div class="card" onclick="navigateOpenPage('${type}', '${item.id}')">
            <img src="${item.cover_url || ''}" class="c-img">
            <div class="c-title">${escapeHtml(item.title || 'Untitled')}</div>
            <div class="c-sub">${escapeHtml(item.artist || item.creator || '')}</div>
        </div>
    `;
}

function renderLibrary() {
    setActiveNav('library');
    const main = document.getElementById('main-view');

    let html = `<span class="section-h">Library</span>`;

    if (state.user) {
        html += `<div style="display:flex;gap:10px;margin-bottom:18px;overflow-x:auto;padding-bottom:4px;">`;
        if (state.user.role === 'artist') {
            html += `
                <button class="submit-btn" style="margin:0;width:auto;padding:10px 18px;" onclick="navigateCreateAlbum()">New Album</button>
                <button class="submit-btn" style="margin:0;width:auto;padding:10px 18px;" onclick="navigateUploadTrack()">Upload Track</button>
            `;
        }
        html += `<button class="submit-btn" style="margin:0;width:auto;padding:10px 18px;" onclick="openNewPlaylistModal()">New Playlist</button>`;
        html += `</div>`;
    }

    const myPlaylists = state.user ? state.playlists.filter(p => p.creator_id === state.user.id) : [];
    html += `<div class="grid-cards">${myPlaylists.map(p => card(p, 'playlist')).join('')}</div>`;

    if (!state.user) html += `<p style="opacity:0.7;font-weight:950;margin-top:14px;">Sign in to manage your library.</p>`;

    main.innerHTML = html;
    closeAll();
    lucide.createIcons();
}

function openNewPlaylistModal() {
    if (!state.user) {
        openModal('auth-modal');
        return;
    }
    document.getElementById('add-modal-title').innerText = "New Playlist";
    const list = document.getElementById('user-playlists-list');
    list.innerHTML = `
        <form onsubmit="handleCreatePlaylist(event)" class="modal-form">
            <input type="text" name="title" placeholder="Playlist Name" required class="glass-input">
            <label class="file-box">
                <span id="lbl-pcover">Cover Image</span>
                <input type="file" name="cover" accept="image/*" required onchange="fileCheck(this, 'lbl-pcover')">
            </label>
            <button class="glass-primary" style="height:56px;border-radius:20px;font-weight:1100;cursor:pointer;">Create</button>
        </form>
    `;
    openModal('add-modal');
    lucide.createIcons();
}

function renderCreateAlbum() {
    setActiveNav('library');
    const main = document.getElementById('main-view');

    if (!state.user || state.user.role !== 'artist') {
        main.innerHTML = `
            <div class="page-topbar">
                <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                <div style="width:1px"></div>
            </div>
            <p style="opacity:0.75;font-weight:950;">Artists only.</p>
        `;
        lucide.createIcons();
        closeAll();
        return;
    }

    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <span class="section-h" style="margin-top:8px;">New Album</span>
        <div class="upload-container">
            <form onsubmit="handleCreateAlbum(event)" class="modal-form">
                <input type="text" name="title" placeholder="Album Title" required class="upload-field">
                <label class="file-box">
                    <span id="lbl-acover">Cover Art</span>
                    <input type="file" name="cover" accept="image/*" required onchange="fileCheck(this, 'lbl-acover')">
                </label>
                <button type="submit" class="submit-btn">Create Album</button>
            </form>
        </div>
    `;
    lucide.createIcons();
    closeAll();
}

function renderUploadTrack() {
    setActiveNav('library');
    const main = document.getElementById('main-view');

    if (!state.user || state.user.role !== 'artist') {
        main.innerHTML = `
            <div class="page-topbar">
                <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                <div style="width:1px"></div>
            </div>
            <p style="opacity:0.75;font-weight:950;">Artists only.</p>
        `;
        lucide.createIcons();
        closeAll();
        return;
    }

    fetch(`${API_URL}/artist-albums?artist_id=${state.user.id}`)
        .then(r => r.json())
        .then(albums => {
            let albumOpts = `<option value="single">Single (No Album)</option>`;
            (albums || []).forEach(a => albumOpts += `<option value="${a.id}">${escapeHtml(a.title || '')}</option>`);

            main.innerHTML = `
                <div class="page-topbar">
                    <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                    <div style="width:1px"></div>
                </div>

                <span class="section-h" style="margin-top:8px;">Upload Track</span>
                <div class="upload-container">
                    <form onsubmit="handleUploadTrack(event)" class="modal-form">
                        <select name="album_id" class="upload-field">${albumOpts}</select>
                        <input type="text" name="title" placeholder="Track Title" required class="upload-field">

                        <label class="file-box">
                            <span id="lbl-cover">Cover Art (If Single)</span>
                            <input type="file" name="cover" accept="image/*" onchange="fileCheck(this, 'lbl-cover')">
                        </label>

                        <label class="file-box">
                            <span id="lbl-audio">Audio File</span>
                            <input type="file" name="audio" accept="audio/*" required onchange="fileCheck(this, 'lbl-audio')">
                        </label>

                        <textarea name="lyrics" placeholder="Lyrics or timed: [01:04] line..." rows="6" class="upload-field"></textarea>
                        <button type="submit" class="submit-btn">Upload</button>
                    </form>
                </div>
            `;
            lucide.createIcons();
            closeAll();
        })
        .catch(() => {
            main.innerHTML = `
                <div class="page-topbar">
                    <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                    <div style="width:1px"></div>
                </div>
                <p style="opacity:0.75;font-weight:950;">Failed to load albums.</p>
            `;
            lucide.createIcons();
            closeAll();
        });
}

function renderSearchPage() {
    setActiveNav('search');
    const main = document.getElementById('main-view');
    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <span class="section-h" style="margin-top:8px;">Search</span>
        <input type="text" id="search-input" placeholder="Search songs..." class="search-bar" oninput="handleSearch(this.value)">
        <div id="search-results" style="margin-top:12px;"></div>
    `;
    lucide.createIcons();
    closeAll();
}

function renderSunoToolsPage() {
    setActiveNav('suno');
    const main = document.getElementById('main-view');
    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <span class="section-h" style="margin-top:8px;">Lyrics & Suno</span>
        <div class="upload-container">
            <div style="display:flex;gap:10px;flex-wrap:wrap;margin-bottom:12px;">
                <a href="https://suno.ai" target="_blank" class="submit-btn" style="margin:0;width:auto;padding:10px 18px;text-decoration:none;display:inline-flex;align-items:center;gap:8px;">
                    <i data-lucide="external-link" style="width:18px;"></i>
                    Open Suno.ai
                </a>
            </div>

            <form onsubmit="handleGenerateLyrics(event)" class="modal-form">
                <select name="genre" class="upload-field" required>
                    <option value="" disabled selected>Select genre</option>
                    <option>Pop</option>
                    <option>Hip-Hop</option>
                    <option>Rap</option>
                    <option>R&B</option>
                    <option>Rock</option>
                    <option>EDM</option>
                    <option>Drill</option>
                    <option>Lo-fi</option>
                    <option>Afrobeat</option>
                    <option>Reggaeton</option>
                </select>

                <textarea name="about" rows="6" class="upload-field" placeholder="About the song (theme, mood, story, language, hooks)..." required></textarea>

                <button class="submit-btn" type="submit" id="gen-btn">Generate Lyrics</button>
            </form>

            <div style="margin-top:15px;">
                <div style="display:flex;justify-content:space-between;align-items:center;gap:10px;margin-bottom:10px;">
                    <div style="font-weight:1100;">Result</div>
                    <button class="submit-btn" style="margin:0;width:auto;padding:10px 16px;" onclick="copyGeneratedLyrics()">Copy</button>
                </div>
                <textarea id="gen-output" class="upload-field" rows="10" placeholder="Generated lyrics will appear here..."></textarea>

                <div style="display:flex;gap:10px;margin-top:10px;">
                    <button class="submit-btn" style="margin:0;flex:1;padding:14px;" onclick="openSunoWithGuide()">Go to Suno</button>
                </div>
            </div>
        </div>
    `;
    lucide.createIcons();
    closeAll();
}

async function handleGenerateLyrics(e) {
    e.preventDefault();
    const btn = document.getElementById('gen-btn');
    const out = document.getElementById('gen-output');
    const fd = new FormData(e.target);

    const payload = {
        genre: (fd.get('genre') || '').toString(),
        about: (fd.get('about') || '').toString(),
        language: 'ru'
    };

    btn.disabled = true;
    btn.style.opacity = "0.75";
    btn.innerText = "Generating...";

    try {
        const res = await fetch(`${API_URL}/generate-lyrics`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (res.ok) {
            const data = await res.json();
            out.value = data.lyrics || '';
        } else {
            const txt = await res.text();
            out.value = (txt || 'AI error').replace(/["\n]/g, '');
        }
    } catch (err) {
        out.value = 'Connection error';
    } finally {
        btn.disabled = false;
        btn.style.opacity = "1";
        btn.innerText = "Generate Lyrics";
    }
}

function copyGeneratedLyrics() {
    const out = document.getElementById('gen-output');
    if (!out) return;
    out.select();
    out.setSelectionRange(0, 999999);
    document.execCommand('copy');
}

function openSunoWithGuide() {
    window.open('https://suno.ai', '_blank');
}

function renderEditProfile() {
    setActiveNav('library');
    const main = document.getElementById('main-view');

    if (!state.user) {
        main.innerHTML = `
            <div class="page-topbar">
                <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                <div style="width:1px"></div>
            </div>
            <p style="opacity:0.75;font-weight:950;">Sign in first.</p>
        `;
        lucide.createIcons();
        closeAll();
        return;
    }

    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <span class="section-h" style="margin-top:8px;">Edit Profile</span>
        <div class="upload-container">
            <form onsubmit="handleUpdateProfile(event)" class="modal-form">
                <input type="text" name="name" value="${escapeAttr(state.user.name || '')}" placeholder="Name" class="upload-field">
                <textarea name="bio" placeholder="Bio" rows="4" class="upload-field">${escapeHtml(state.user.bio || '')}</textarea>
                <button type="submit" class="submit-btn">Save</button>
            </form>
        </div>
    `;
    lucide.createIcons();
    closeAll();
}

function fileCheck(input, id) {
    if (input.files && input.files[0]) {
        const el = document.getElementById(id);
        if (el) el.innerText = input.files[0].name;
    }
}

async function handleCreateAlbum(e) {
    e.preventDefault();
    const fd = new FormData(e.target);
    fd.append('artist_id', state.user.id);
    fd.append('artist_name', state.user.name);
    await fetch(`${API_URL}/create-album`, { method: 'POST', body: fd });
    await fetchContent();
    navigateLibrary();
}

async function handleUploadTrack(e) {
    e.preventDefault();
    const fd = new FormData(e.target);
    fd.append('artist_id', state.user.id);
    fd.append('artist_name', state.user.name);
    await fetch(`${API_URL}/upload-track`, { method: 'POST', body: fd });
    await fetchContent();
    navigateHome();
}

async function handleCreatePlaylist(e) {
    e.preventDefault();
    const fd = new FormData(e.target);
    fd.append('creator', state.user.name);
    fd.append('creator_id', state.user.id);
    await fetch(`${API_URL}/create-playlist`, { method: 'POST', body: fd });
    closeModal('add-modal');
    await fetchContent();
    navigateLibrary();
}

async function handleUpdateProfile(e) {
    e.preventDefault();
    const fd = new FormData(e.target);
    const data = { id: state.user.id, name: fd.get('name'), bio: fd.get('bio') };

    const res = await fetch(`${API_URL}/update-profile`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    const u = await res.json();
    localStorage.setItem('user', JSON.stringify(u));
    state.user = u;
    setupUI();
    navigateHome();
}

async function deleteTrack(id) {
    if (!confirm("Delete this track?")) return;
    await fetch(`${API_URL}/delete-track`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id })
    });
    await fetchContent();
    navigateLibrary();
}

async function handleSearch(q) {
    const container = document.getElementById('search-results');
    if (!container) return;

    const s = (q || '').trim();
    if (s.length < 2) {
        container.innerHTML = '';
        return;
    }

    const res = await fetch(`${API_URL}/search?q=${encodeURIComponent(s)}`);
    const tracks = await res.json();

    container.innerHTML = (tracks || []).map((t, i) => `
        <div class="playlist-item-add" style="justify-content:space-between;">
            <div onclick="playQueue(${i}, '${encodeURIComponent(JSON.stringify(tracks))}')" style="flex:1;display:flex;align-items:center;gap:12px;cursor:pointer;min-width:0;">
                <img src="${t.cover_url || ''}" style="width:46px;height:46px;border-radius:16px;object-fit:cover;border:1px solid var(--stroke);">
                <div style="min-width:0;">
                    <div style="font-weight:1100;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHtml(t.title || '')}</div>
                    <div style="font-size:12px;font-weight:900;color:var(--muted2);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHtml(t.artist || '')}</div>
                </div>
            </div>
            <button class="icon-btn" onclick="openAddToPlaylistModal('${t.id}')" style="width:44px;height:44px;border-radius:16px;"><i data-lucide="list-plus"></i></button>
        </div>
    `).join('');

    lucide.createIcons();
}

async function openPage(type, id, fromRoute = false) {
    const res = await fetch(`${API_URL}/${type}?id=${id}`);
    const data = await res.json();
    const info = data.info || {};
    const tracks = data.tracks || [];

    if (!fromRoute) {
        navigateOpenPage(type, id);
        return;
    }

    const main = document.getElementById('main-view');
    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <div class="upload-container" style="margin-top:6px;text-align:center;">
            <img src="${info.cover_url || ''}" style="width:200px;height:200px;border-radius:26px;box-shadow:0 34px 120px rgba(0,0,0,0.55);object-fit:cover;border:1px solid var(--stroke);">
            <h2 style="margin:16px 0 0;font-size:24px;line-height:1.2;font-weight:1150;">${escapeHtml(info.title || '')}</h2>
            <p style="margin:6px 0 0;font-weight:900;color:var(--muted);">${escapeHtml(info.artist || info.creator || '')}</p>
            <p style="margin:6px 0 0;font-size:12px;font-weight:900;color:var(--muted2);">${type === 'album' ? (info.is_single ? 'Single' : 'Album') : 'Playlist'}</p>
        </div>

        <div class="upload-container" style="margin-top:14px;padding:12px;">
            ${tracks.map((t, i) => {
        const owner = state.user && (state.user.id === info.artist_id || state.user.id === info.creator_id);
        const editBtn = owner && state.user && state.user.role === 'artist'
            ? `<button class="icon-btn" onclick="navigateEditLyrics('${t.id}', '${encodeURIComponent((t.lyrics||"").toString())}', '${encodeURIComponent((t.title||"").toString())}', '${encodeURIComponent((t.artist||"").toString())}')" style="width:44px;height:44px;border-radius:16px;"><i data-lucide="file-pen-line"></i></button>`
            : '';
        return `
                    <div class="playlist-item-add" style="padding:12px;justify-content:space-between;">
                        <div onclick="playQueue(${i}, '${encodeURIComponent(JSON.stringify(tracks))}')" style="flex:1;display:flex;align-items:center;gap:12px;cursor:pointer;min-width:0;">
                            <span style="width:26px;text-align:left;font-weight:950;color:var(--muted2);">${i + 1}</span>
                            <div style="min-width:0;">
                                <div style="font-weight:1100;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHtml(t.title || '')}</div>
                                <div style="font-size:12px;font-weight:900;color:var(--muted2);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${escapeHtml(t.artist || '')}</div>
                            </div>
                        </div>
                        <div style="display:flex;gap:10px;align-items:center;">
                            ${owner ? `<button onclick="deleteTrack('${t.id}')" class="del-btn"><i data-lucide="trash-2"></i></button>` : ''}
                            ${editBtn}
                            <button class="icon-btn" onclick="openAddToPlaylistModal('${t.id}')" style="width:44px;height:44px;border-radius:16px;"><i data-lucide="list-plus"></i></button>
                        </div>
                    </div>
                `;
    }).join('')}
        </div>
    `;
    lucide.createIcons();
    closeAll();
}

function renderEditLyrics(s) {
    setActiveNav('library');
    const main = document.getElementById('main-view');

    if (!state.user || state.user.role !== 'artist') {
        main.innerHTML = `
            <div class="page-topbar">
                <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
                <div style="width:1px"></div>
            </div>
            <p style="opacity:0.75;font-weight:950;">Artists only.</p>
        `;
        lucide.createIcons();
        closeAll();
        return;
    }

    const trackId = s.trackId;
    const title = decodeURIComponent(s.title || '');
    const artist = decodeURIComponent(s.artist || '');
    const existingLyrics = decodeURIComponent(s.existingLyrics || '');

    main.innerHTML = `
        <div class="page-topbar">
            <button class="page-back" onclick="goBack()"><i data-lucide="arrow-left"></i> Back</button>
            <div style="width:1px"></div>
        </div>

        <span class="section-h" style="margin-top:8px;">Edit Lyrics</span>

        <div class="upload-container lyr-editor">
            <div class="lyr-top">
                <div class="pill">Track: <b>${escapeHtml(title || 'Untitled')}</b></div>
                <div class="pill">Artist: <b>${escapeHtml(artist || '')}</b></div>
                <div class="pill">Mode: <b id="tap-mode">TAP</b></div>
            </div>

            <div class="pill" style="text-align:left;">
                Play the song, then press <b>Space</b> to timestamp the next line. Click a line to jump audio to that timestamp.
            </div>

            <div class="lyr-preview" id="lyr-lines"></div>

            <textarea id="lyr-raw" rows="10" class="upload-field" placeholder="Write lyrics here, one line per row...">${escapeHtml(stripTime(existingLyrics))}</textarea>

            <div style="display:flex;gap:10px;flex-wrap:wrap;">
                <button class="submit-btn" style="flex:1;min-width:180px;" onclick="startTap('${trackId}')">Start Tap (Space)</button>
                <button class="glass-ghost" style="flex:1;min-width:180px;" onclick="stopTap()">Stop Tap</button>
                <button class="glass-ghost" style="flex:1;min-width:180px;" onclick="buildTimed()">Build Timed Lyrics</button>
                <button class="submit-btn" style="flex:1;min-width:180px;" onclick="saveTimedLyrics('${trackId}')">Save</button>
            </div>

            <textarea id="lyr-timed" rows="10" class="upload-field" placeholder="Timed lyrics output will appear here...">${escapeHtml(normalizeTimed(existingLyrics))}</textarea>
        </div>
    `;

    lucide.createIcons();
    closeAll();

    setTimeout(() => {
        renderTapPreview();
    }, 0);
}

function playQueue(idx, tracksStr) {
    state.queue = JSON.parse(decodeURIComponent(tracksStr)) || [];
    state.currentIndex = idx;
    loadTrack(state.queue[idx]);
}

function loadTrack(t) {
    if (!t) return;
    state.currentTrackId = t.id;
    state.lastOpenedTrack = t;
    audio.src = t.audio_url;
    audio.play();
    state.isPlaying = true;
    updatePlayerUI(t);
    parseLyrics(t.lyrics);
    const mp = document.getElementById('mini-player');
    if (mp) mp.classList.remove('hidden');
}

function parseLyrics(text) {
    if (!text) {
        state.isKaraoke = false;
        state.lyricsData = [];
        return;
    }

    const lines = text.split('\n');
    const parsed = [];
    let hasTime = false;

    lines.forEach(line => {
        const match = line.match(/\[(\d+):(\d+)\](.*)/);
        if (match) {
            hasTime = true;
            const mm = parseInt(match[1], 10);
            const ss = parseInt(match[2], 10);
            const content = (match[3] || '').trim();
            const time = (mm * 60) + ss;
            parsed.push({ time, text: content });
        }
    });

    if (hasTime) {
        state.isKaraoke = true;
        state.lyricsData = parsed.sort((a, b) => a.time - b.time);
    } else {
        state.isKaraoke = false;
        state.lyricsData = text;
    }

    const lo = document.getElementById('lyrics-overlay');
    if (lo && !lo.classList.contains('hidden')) renderLyrics();
}

function updatePlayerUI(t) {
    const mpCover = document.getElementById('mp-cover');
    const fpCover = document.getElementById('fp-cover');
    const fpBg = document.getElementById('fp-bg');
    const lyricBg = document.getElementById('lyric-bg');

    if (mpCover) mpCover.src = t.cover_url || '';
    if (fpCover) fpCover.src = t.cover_url || '';
    if (fpBg) fpBg.src = t.cover_url || '';
    if (lyricBg) lyricBg.src = t.cover_url || '';

    const mpTitle = document.getElementById('mp-title');
    const mpArtist = document.getElementById('mp-artist');
    const fpTitle = document.getElementById('fp-title');
    const fpArtist = document.getElementById('fp-artist');

    if (mpTitle) mpTitle.innerText = t.title || '';
    if (mpArtist) mpArtist.innerText = t.artist || '';
    if (fpTitle) fpTitle.innerText = t.title || '';
    if (fpArtist) fpArtist.innerText = t.artist || '';

    const icon = state.isPlaying ? 'pause' : 'play';

    const mpPlay = document.getElementById('mp-play');
    const fpPlayIcon = document.getElementById('fp-play-icon');

    if (mpPlay) mpPlay.innerHTML = `<i data-lucide="${icon}"></i>`;
    if (fpPlayIcon) fpPlayIcon.setAttribute('data-lucide', icon);

    const loopBtn = document.getElementById('fp-loop');
    if (loopBtn) {
        loopBtn.style.opacity = audio.loop ? '1' : '0.85';
        loopBtn.style.borderColor = audio.loop ? 'rgba(125,211,252,0.45)' : 'var(--stroke)';
    }

    lucide.createIcons();
}

function renderLyrics() {
    const container = document.getElementById('lyrics-container');
    if (!container) return;
    container.innerHTML = '';

    const wrap = document.createElement('div');
    wrap.className = 'lyrics-wrap';
    container.appendChild(wrap);

    if (state.isKaraoke) {
        const spacerTop = document.createElement('div');
        spacerTop.style.height = "35vh";
        wrap.appendChild(spacerTop);

        state.lyricsData.forEach((line, index) => {
            const div = document.createElement('div');
            div.className = 'karaoke-line';
            div.id = `line-${index}`;
            div.innerText = (line.text || "♫").trim();
            div.onclick = () => { audio.currentTime = line.time; };
            wrap.appendChild(div);
        });

        const spacerBot = document.createElement('div');
        spacerBot.style.height = "35vh";
        wrap.appendChild(spacerBot);
        return;
    }

    const raw = (state.lyricsData || "").toString().replaceAll('\r\n', '\n');
    if (!raw.trim()) {
        const empty = document.createElement('div');
        empty.className = 'static-lyrics';
        empty.innerText = 'No Lyrics Available';
        wrap.appendChild(empty);
        return;
    }

    raw.split('\n').forEach(l => {
        const s = (l || '').trim();
        if (!s) return;

        if (/^\[.*\]$/.test(s)) {
            const tag = document.createElement('div');
            tag.className = 'lyrics-tag';
            tag.innerText = s.replace('[', '').replace(']', '');
            wrap.appendChild(tag);
        } else {
            const line = document.createElement('div');
            line.className = 'lyrics-line';
            line.innerText = s;
            wrap.appendChild(line);
        }
    });
}

function syncLyrics() {
    const lo = document.getElementById('lyrics-overlay');
    if (!state.isKaraoke || !lo || lo.classList.contains('hidden')) return;

    const time = audio.currentTime;
    let activeIdx = -1;

    for (let i = 0; i < state.lyricsData.length; i++) {
        if (time >= state.lyricsData[i].time) activeIdx = i;
        else break;
    }

    if (activeIdx !== -1) {
        document.querySelectorAll('.karaoke-line').forEach(el => el.classList.remove('active'));
        const activeEl = document.getElementById(`line-${activeIdx}`);
        if (activeEl) {
            activeEl.classList.add('active');
            activeEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }
}

function updateProgress() {
    const pct = (audio.currentTime / audio.duration) * 100 || 0;
    const seek = document.getElementById('fp-seek');
    const fill = document.getElementById('mp-fill');
    if (fill) fill.style.width = pct + '%';

    if (seek) {
        seek.value = pct;
        seek.style.background = `linear-gradient(to right, rgba(125,211,252,0.98) ${pct}%, rgba(255,255,255,0.18) ${pct}%)`;
    }

    const curr = document.getElementById('fp-curr');
    const dur = document.getElementById('fp-dur');
    if (curr) curr.innerText = fmtTime(audio.currentTime);
    if (dur) dur.innerText = fmtTime(audio.duration || 0);

    syncLyrics();
}

function togglePlay() {
    if (!audio.src) return;

    if (state.isPlaying) {
        audio.pause();
        state.isPlaying = false;
    } else {
        audio.play();
        state.isPlaying = true;
    }

    const icon = state.isPlaying ? 'pause' : 'play';
    const mpPlay = document.getElementById('mp-play');
    const fpPlayIcon = document.getElementById('fp-play-icon');
    if (mpPlay) mpPlay.innerHTML = `<i data-lucide="${icon}"></i>`;
    if (fpPlayIcon) fpPlayIcon.setAttribute('data-lucide', icon);
    lucide.createIcons();
}

function nextTrack() {
    if (!state.queue || state.queue.length === 0) return;

    if (state.currentIndex < state.queue.length - 1) {
        state.currentIndex++;
        loadTrack(state.queue[state.currentIndex]);
        return;
    }

    if (!audio.loop) {
        state.currentIndex = 0;
        loadTrack(state.queue[state.currentIndex]);
    }
}

function prevTrack() {
    if (!state.queue || state.queue.length === 0) return;

    if (audio.currentTime > 3) {
        audio.currentTime = 0;
        return;
    }

    if (state.currentIndex > 0) {
        state.currentIndex--;
        loadTrack(state.queue[state.currentIndex]);
    } else {
        state.currentIndex = state.queue.length - 1;
        loadTrack(state.queue[state.currentIndex]);
    }
}

function toggleLoop() {
    audio.loop = !audio.loop;
    const loopBtn = document.getElementById('fp-loop');
    if (loopBtn) {
        loopBtn.style.opacity = audio.loop ? '1' : '0.85';
        loopBtn.style.borderColor = audio.loop ? 'rgba(125,211,252,0.45)' : 'var(--stroke)';
    }
}

function setupAudio() {
    audio.addEventListener('timeupdate', updateProgress);

    const seek = document.getElementById('fp-seek');
    if (seek) {
        seek.oninput = (e) => {
            if (!audio.duration) return;
            audio.currentTime = (e.target.value / 100) * audio.duration;
        };
    }

    audio.addEventListener('ended', () => {
        if (audio.loop) return;
        nextTrack();
    });
}

function fmtTime(s) {
    const m = Math.floor(s / 60);
    const sec = Math.floor(s % 60);
    return `${m}:${sec < 10 ? '0' : ''}${sec}`;
}

function openFullPlayer() {
    const fp = document.getElementById('full-player');
    if (fp) fp.classList.remove('hidden');
}

function closeFullPlayer() {
    const fp = document.getElementById('full-player');
    if (fp) fp.classList.add('hidden');
}

function toggleLyricsPage() {
    const fp = document.getElementById('full-player');
    const lp = document.getElementById('lyrics-overlay');
    if (!fp || !lp) return;

    if (lp.classList.contains('hidden')) {
        renderLyrics();
        fp.classList.add('hidden');
        lp.classList.remove('hidden');
    } else {
        lp.classList.add('hidden');
        fp.classList.remove('hidden');
    }
}

function closeAll() {
    const fp = document.getElementById('full-player');
    const lp = document.getElementById('lyrics-overlay');
    if (fp) fp.classList.add('hidden');
    if (lp) lp.classList.add('hidden');
}

async function openAddToPlaylistModal(trackId) {
    if (!state.user) {
        openModal('auth-modal');
        return;
    }
    state.currentTrackId = trackId;

    const res = await fetch(`${API_URL}/user-playlists?user_id=${state.user.id}`);
    const playlists = await res.json();

    document.getElementById('add-modal-title').innerText = "Add to Playlist";
    const list = document.getElementById('user-playlists-list');

    if (!playlists || playlists.length === 0) {
        list.innerHTML = `<div style="opacity:0.8;font-weight:950;padding:8px 2px;">No playlists found. Create one first.</div>`;
        openModal('add-modal');
        return;
    }

    list.innerHTML = playlists.map(p => `
        <div onclick="addToPlaylist('${p.id}')" class="playlist-item-add">
            <img src="${p.cover_url || ''}">
            <span style="font-weight:1100;">${escapeHtml(p.title || 'Untitled')}</span>
        </div>
    `).join('');

    openModal('add-modal');
    lucide.createIcons();
}

async function addToPlaylist(playlistId) {
    await fetch(`${API_URL}/add-to-playlist`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ playlist_id: playlistId, track_id: state.currentTrackId })
    });
    closeModal('add-modal');
}

function openAddToPlaylist() {
    if (state.queue[state.currentIndex]) openAddToPlaylistModal(state.queue[state.currentIndex].id);
}

async function openStats() {
    const res = await fetch(`${API_URL}/stats`);
    const data = await res.json();
    const content = document.getElementById('stats-content');

    const tracks = data.tracks ?? 0;
    const users = data.users ?? 0;
    const top = data.top_artists ?? [];

    content.innerHTML = `
        <div class="stats-row">
            <div class="stat-card">
                <div class="stat-k">Total Tracks</div>
                <div class="stat-v">${tracks}</div>
            </div>
            <div class="stat-card">
                <div class="stat-k">Community Users</div>
                <div class="stat-v">${users}</div>
            </div>
        </div>

        <div class="stats-title">Top 5 Artists</div>
        <div class="stats-list">
            ${top.map((a, i) => `
                <div class="stats-item">
                    <span>${i + 1}. ${escapeHtml(a._id || '')}</span>
                    <b>${a.count || 0} tracks</b>
                </div>
            `).join('')}
        </div>
    `;

    openModal('stats-modal');
}

async function doAuth(type) {
    const form = document.getElementById('auth-form');
    const data = Object.fromEntries(new FormData(form));
    const errorBox = document.getElementById('auth-error');

    errorBox.classList.add('hidden');
    errorBox.innerText = '';

    try {
        const res = await fetch(`${API_URL}/${type}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (res.ok) {
            const u = await res.json();
            localStorage.setItem('user', JSON.stringify(u));
            location.reload();
        } else {
            const txt = await res.text();
            errorBox.innerText = (txt || 'Auth error').replace(/["\n]/g, '');
            errorBox.classList.remove('hidden');
        }
    } catch (e) {
        errorBox.innerText = "Connection error";
        errorBox.classList.remove('hidden');
    }
}

function logout() {
    localStorage.removeItem('user');
    location.reload();
}

function openModal(id) {
    const el = document.getElementById(id);
    if (el) el.classList.remove('hidden');
    lucide.createIcons();
}

function closeModal(id) {
    const el = document.getElementById(id);
    if (el) el.classList.add('hidden');
}

function escapeHtml(s) {
    return (s ?? '').toString()
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", "&#039;");
}

function escapeAttr(s) {
    return escapeHtml(s).replaceAll('\n', ' ');
}

function stripTime(txt) {
    const t = (txt || '').toString().split('\n');
    const out = [];
    for (const line of t) {
        const m = line.match(/^\[(\d+):(\d+)\]\s*(.*)$/);
        out.push(m ? (m[3] || '') : line);
    }
    return out.join('\n').trim();
}

function normalizeTimed(txt) {
    const t = (txt || '').toString().trim();
    if (!t) return '';
    const lines = t.split('\n').map(x => x.trim()).filter(Boolean);
    const has = lines.some(l => /^\[\d+:\d+\]/.test(l));
    if (has) return lines.join('\n');
    return '';
}

function parsePlainLines() {
    const raw = document.getElementById('lyr-raw');
    const lines = (raw?.value || '').split('\n').map(x => x.trim()).filter(Boolean);
    return lines;
}

function renderTapPreview() {
    const box = document.getElementById('lyr-lines');
    if (!box) return;

    const lines = parsePlainLines();
    tapState.lines = lines;

    if (!tapState.times || tapState.times.length !== lines.length) {
        tapState.times = new Array(lines.length).fill(null);
    }

    box.innerHTML = lines.map((l, i) => {
        const t = tapState.times[i];
        const label = t == null ? '--:--' : `[${fmtTimeTag(t)}]`;
        const active = tapState.on && i === tapState.idx ? ' active' : '';
        return `<div class="lyr-line${active}" onclick="jumpToTap(${i})">${label} ${escapeHtml(l)}</div>`;
    }).join('');
}

function fmtTimeTag(sec) {
    sec = Math.max(0, Math.floor(sec));
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m < 10 ? '0' : ''}${m}:${s < 10 ? '0' : ''}${s}`;
}

function jumpToTap(i) {
    if (!tapState.times[i] && tapState.on) return;
    if (tapState.times[i] != null) audio.currentTime = tapState.times[i];
    tapState.idx = i;
    renderTapPreview();
}

function startTap(trackId) {
    if (!audio.src) {
        alert('Open the track and press play first.');
        return;
    }
    tapState.on = true;
    tapState.idx = 0;
    tapState.lines = parsePlainLines();
    tapState.times = new Array(tapState.lines.length).fill(null);

    const mode = document.getElementById('tap-mode');
    if (mode) mode.innerText = 'TAP';

    renderTapPreview();
    document.addEventListener('keydown', tapKeyHandler);
}

function stopTap() {
    tapState.on = false;
    const mode = document.getElementById('tap-mode');
    if (mode) mode.innerText = 'STOP';
    document.removeEventListener('keydown', tapKeyHandler);
    renderTapPreview();
}

function tapKeyHandler(e) {
    if (!tapState.on) return;
    if (e.code !== 'Space') return;
    e.preventDefault();

    const i = tapState.idx;
    if (!tapState.lines || i >= tapState.lines.length) return;

    tapState.times[i] = audio.currentTime;
    tapState.idx = Math.min(i + 1, tapState.lines.length - 1);

    renderTapPreview();
    buildTimed();
}

function buildTimed() {
    const out = document.getElementById('lyr-timed');
    if (!out) return;

    const lines = parsePlainLines();
    if (!tapState.times || tapState.times.length !== lines.length) {
        tapState.times = new Array(lines.length).fill(null);
    }

    const timed = lines.map((l, i) => {
        const t = tapState.times[i];
        if (t == null) return l;
        return `[${fmtTimeTag(t)}] ${l}`;
    }).join('\n');

    out.value = timed.trim();
    renderTapPreview();
}

async function saveTimedLyrics(trackId) {
    if (!trackId) return;

    if (!state.user) {
        alert('Sign in first');
        return;
    }

    const timed = document.getElementById('lyr-timed')?.value || '';
    const payload = { track_id: trackId, user_id: state.user.id, lyrics: timed };

    const res = await fetch(`${API_URL}/update-lyrics`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });

    if (!res.ok) {
        const t = await res.text();
        alert((t || 'Save failed').replace(/["\n]/g, ''));
        return;
    }

    await fetchContent();
    alert('Saved');
    goBack();
}
