// Token Visualizer Interactive UI
(function() {
    'use strict';

    // State
    let currentView = 'text'; // 'text' or 'table'
    let currentTheme = 'dark'; // 'dark' or 'light'
    let selectedTokenIndex = null;
    let tokens = [];

    // DOM Elements
    const textView = document.getElementById('text-view');
    const tableView = document.getElementById('table-view');
    const viewToggleBtn = document.getElementById('view-toggle');
    const viewLabel = document.getElementById('view-label');
    const viewIcon = document.getElementById('view-icon');
    const themeToggleBtn = document.getElementById('theme-toggle');
    const themeLabel = document.getElementById('theme-label');
    const themeIcon = document.getElementById('theme-icon');
    const searchInput = document.getElementById('search-input');
    const searchCount = document.getElementById('search-count');
    const helpModal = document.getElementById('help-modal');
    const closeHelpBtn = document.getElementById('close-help');

    // Initialize
    function init() {
        // Collect all tokens
        tokens = Array.from(document.querySelectorAll('.token'));

        // Calculate statistics
        calculateStatistics();

        // Load theme preference
        loadTheme();

        // Setup event listeners
        setupEventListeners();

        // Initial focus on body (not search)
        document.body.focus();
    }

    // Calculate token statistics
    function calculateStatistics() {
        const lengths = tokens.map(token => parseInt(token.dataset.length));
        const avgLength = lengths.reduce((a, b) => a + b, 0) / lengths.length;
        const maxLength = Math.max(...lengths);
        const minLength = Math.min(...lengths);

        document.getElementById('stat-avg').textContent = avgLength.toFixed(1);
        document.getElementById('stat-max').textContent = maxLength;
        document.getElementById('stat-min').textContent = minLength;
    }

    // Setup event listeners
    function setupEventListeners() {
        // View toggle
        viewToggleBtn.addEventListener('click', toggleView);

        // Theme toggle
        themeToggleBtn.addEventListener('click', toggleTheme);

        // Search functionality
        searchInput.addEventListener('input', handleSearch);

        // Token click handlers (text view)
        tokens.forEach((token, index) => {
            token.addEventListener('click', () => selectToken(index));
        });

        // Table row click handlers
        const tableRows = document.querySelectorAll('.token-table tbody tr');
        tableRows.forEach((row, index) => {
            row.addEventListener('click', () => selectToken(index));
        });

        // Copy button handlers
        const copyButtons = document.querySelectorAll('.btn-copy');
        copyButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                copyToClipboard(btn.dataset.text);
                showCopyFeedback(btn);
            });
        });

        // Help modal
        closeHelpBtn.addEventListener('click', closeHelp);

        // Keyboard shortcuts
        document.addEventListener('keydown', handleKeyboard);

        // Click outside modal to close
        helpModal.addEventListener('click', (e) => {
            if (e.target === helpModal) {
                closeHelp();
            }
        });
    }

    // Toggle view between text and table
    function toggleView() {
        if (currentView === 'text') {
            currentView = 'table';
            textView.classList.remove('active');
            tableView.classList.add('active');
            viewLabel.textContent = 'Switch to Text';
            viewIcon.textContent = '📊';
        } else {
            currentView = 'text';
            tableView.classList.remove('active');
            textView.classList.add('active');
            viewLabel.textContent = 'Switch to Table';
            viewIcon.textContent = '📝';
        }
    }

    // Toggle theme between dark and light
    function toggleTheme() {
        if (currentTheme === 'dark') {
            currentTheme = 'light';
            document.body.classList.add('light-theme');
            themeLabel.textContent = 'Light';
            themeIcon.textContent = '☀️';
        } else {
            currentTheme = 'dark';
            document.body.classList.remove('light-theme');
            themeLabel.textContent = 'Dark';
            themeIcon.textContent = '🌙';
        }
        saveTheme();
    }

    // Save theme preference to localStorage
    function saveTheme() {
        localStorage.setItem('cc-token-theme', currentTheme);
    }

    // Load theme preference from localStorage
    function loadTheme() {
        const savedTheme = localStorage.getItem('cc-token-theme');
        if (savedTheme === 'light') {
            toggleTheme();
        }
    }

    // Handle search input
    function handleSearch() {
        const query = searchInput.value.toLowerCase().trim();

        if (query === '') {
            clearSearch();
            return;
        }

        let matchCount = 0;
        const tableRows = document.querySelectorAll('.token-table tbody tr');

        tokens.forEach((token, index) => {
            const text = token.dataset.text.toLowerCase();
            const matches = text.includes(query);

            if (matches) {
                matchCount++;
                token.classList.add('highlighted');
                tableRows[index]?.classList.add('highlighted');
            } else {
                token.classList.remove('highlighted');
                tableRows[index]?.classList.remove('highlighted');
            }
        });

        searchCount.textContent = `${matchCount} match${matchCount !== 1 ? 'es' : ''}`;
    }

    // Clear search
    function clearSearch() {
        searchInput.value = '';
        searchCount.textContent = '';
        tokens.forEach(token => {
            token.classList.remove('highlighted');
        });
        const tableRows = document.querySelectorAll('.token-table tbody tr');
        tableRows.forEach(row => {
            row.classList.remove('highlighted');
        });
        selectedTokenIndex = null;
    }

    // Select a token
    function selectToken(index) {
        // Clear previous selection
        if (selectedTokenIndex !== null) {
            tokens[selectedTokenIndex].classList.remove('highlighted');
            const prevRow = document.querySelector(`.token-table tbody tr[data-index="${selectedTokenIndex}"]`);
            if (prevRow) prevRow.classList.remove('highlighted');
        }

        // Select new token
        selectedTokenIndex = index;
        tokens[index].classList.add('highlighted');
        const row = document.querySelector(`.token-table tbody tr[data-index="${index}"]`);
        if (row) {
            row.classList.add('highlighted');
            // Scroll into view if in table mode
            if (currentView === 'table') {
                row.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        }

        // Scroll token into view if in text mode
        if (currentView === 'text') {
            tokens[index].scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }

    // Copy text to clipboard
    function copyToClipboard(text) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).catch(err => {
                console.error('Failed to copy:', err);
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    }

    // Fallback copy method
    function fallbackCopy(text) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
        } catch (err) {
            console.error('Fallback copy failed:', err);
        }
        document.body.removeChild(textarea);
    }

    // Show copy feedback
    function showCopyFeedback(btn) {
        const originalText = btn.textContent;
        btn.textContent = '✓';
        btn.style.color = '#00FF7F';
        setTimeout(() => {
            btn.textContent = originalText;
            btn.style.color = '';
        }, 1000);
    }

    // Handle keyboard shortcuts
    function handleKeyboard(e) {
        // Don't handle shortcuts when typing in search
        const isSearchFocused = document.activeElement === searchInput;

        // Tab: Toggle view
        if (e.key === 'Tab') {
            e.preventDefault();
            toggleView();
        }

        // /: Focus search (only when not already focused)
        else if (e.key === '/' && !isSearchFocused) {
            e.preventDefault();
            searchInput.focus();
            searchInput.select();
        }

        // Escape: Clear search and deselect
        else if (e.key === 'Escape') {
            e.preventDefault();
            clearSearch();
            searchInput.blur();
            closeHelp();
        }

        // ?: Show help
        else if ((e.key === '?' || e.key === 'h') && !isSearchFocused) {
            e.preventDefault();
            showHelp();
        }

        // t: Toggle theme
        else if (e.key === 't' && !isSearchFocused) {
            e.preventDefault();
            toggleTheme();
        }

        // Arrow keys: Navigate tokens (when not in search)
        else if (!isSearchFocused) {
            if (e.key === 'ArrowDown' || e.key === 'j') {
                e.preventDefault();
                navigateToken(1);
            } else if (e.key === 'ArrowUp' || e.key === 'k') {
                e.preventDefault();
                navigateToken(-1);
            }
        }
    }

    // Navigate to next/previous token
    function navigateToken(direction) {
        let newIndex;
        if (selectedTokenIndex === null) {
            newIndex = direction > 0 ? 0 : tokens.length - 1;
        } else {
            newIndex = selectedTokenIndex + direction;
            if (newIndex < 0) newIndex = 0;
            if (newIndex >= tokens.length) newIndex = tokens.length - 1;
        }
        selectToken(newIndex);
    }

    // Show help modal
    function showHelp() {
        helpModal.classList.add('active');
    }

    // Close help modal
    function closeHelp() {
        helpModal.classList.remove('active');
    }

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();
