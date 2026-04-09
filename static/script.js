// Initialize slider values on page load
function initializeSliderValues() {
    const sliders = document.querySelectorAll('input[type="range"]');
    sliders.forEach(slider => {
        updateSliderValue(slider);
    });
}

// Update slider value display
function updateSliderValue(slider) {
    const container = slider.closest('.slider-container');
    if (container) {
        const valueDisplay = container.querySelector('.slider-value');
        if (valueDisplay) {
            valueDisplay.textContent = slider.value;
        }
    }
}

// Set example mermaid code
function setExample(example) {
    const textarea = document.querySelector('textarea[name="mermaid"]');
    if (textarea) {
        textarea.value = example;
        // Trigger HTMX update by dispatching both input and change events
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
        textarea.dispatchEvent(new Event('change', { bubbles: true }));
        // Also trigger HTMX directly if available
        if (window.htmx) {
            htmx.trigger(textarea, 'changed');
        }
    }
}

// Global terminal state
let term;
let fitAddon;
let lastRenderedContent = '';

function stripAnsi(text) {
    return text.replace(/\u001b\[[0-9;]*m/g, '');
}

// Copy terminal content
function copyTerminalContent() {
    if (!lastRenderedContent) {
        return;
    }

    const btn = document.getElementById('copy-button');
    const originalText = btn.textContent;

    navigator.clipboard.writeText(lastRenderedContent).then(() => {
        btn.textContent = 'Copied!';
        setTimeout(() => {
            btn.textContent = originalText;
        }, 2000);
    });
}

function fitTerminal() {
    if (!term || !fitAddon) return;
    fitAddon.fit();
}

function renderTerminal(content) {
    if (!term) return;

    const normalizedContent = content.replace(/\r\n/g, '\n');
    lastRenderedContent = stripAnsi(normalizedContent);
    term.reset();
    fitTerminal();
    term.write(normalizedContent.replace(/\n/g, '\r\n'));
    term.scrollToTop();
}

// Update terminal theme
function updateTerminalTheme(isDark) {
    if (!term) return;
    
    if (isDark) {
        term.options.theme = {
            background: '#1a1a1a',
            foreground: '#e9ecef',
            cursor: '#e9ecef',
            cursorAccent: '#1a1a1a',
            selection: 'rgba(77, 171, 247, 0.3)',
        };
    } else {
        term.options.theme = {
            background: '#e9ecef',
            foreground: '#212529',
            cursor: '#212529',
            cursorAccent: '#e9ecef',
            selection: 'rgba(0, 123, 255, 0.3)',
        };
    }
}

// Handle HTMX before swap to write to terminal
document.addEventListener('DOMContentLoaded', function() {
    initializeSliderValues();
    
    // Initialize xterm.js terminal
    const terminalContainer = document.getElementById('terminal-container');
    if (terminalContainer && typeof Terminal !== 'undefined') {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark' || 
                       (!document.documentElement.getAttribute('data-theme') && 
                        window.matchMedia('(prefers-color-scheme: dark)').matches);
        
        term = new Terminal({
            fontFamily: "'Roboto Mono', 'Courier New', monospace",
            fontSize: 13,
            lineHeight: 1.15,
            letterSpacing: 0,
            scrollback: 1000,
            disableStdin: true,
            convertEol: true,
            cursorWidth: 0,
            theme: isDark ? {
                background: '#1a1a1a',
                foreground: '#e9ecef',
                cursor: '#e9ecef',
                cursorAccent: '#1a1a1a',
                selection: 'rgba(77, 171, 247, 0.3)',
            } : {
                background: '#e9ecef',
                foreground: '#212529',
                cursor: '#212529',
                cursorAccent: '#e9ecef',
                selection: 'rgba(0, 123, 255, 0.3)',
            }
        });

        fitAddon = new FitAddon.FitAddon();
        term.loadAddon(fitAddon);
        term.open(terminalContainer);
        const resizeObserver = new ResizeObserver(() => {
            fitTerminal();
        });
        resizeObserver.observe(terminalContainer);
        fitTerminal();
    }

    document.body.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.target.id === 'result-code') {
            evt.preventDefault();
            renderTerminal(evt.detail.xhr.responseText);
        }
    });

    // Add loading state to form during HTMX requests
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('htmx:beforeRequest', function() {
            form.classList.add('loading');
        });
        
        form.addEventListener('htmx:afterRequest', function() {
            form.classList.remove('loading');
        });
    }
});

// Add event listeners for sliders
document.addEventListener('DOMContentLoaded', function() {
    const sliders = document.querySelectorAll('input[type="range"]');
    sliders.forEach(slider => {
        slider.addEventListener('input', function() {
            updateSliderValue(this);
        });
    });
});

// Theme management
function getStoredTheme() {
    return localStorage.getItem('theme');
}

function setStoredTheme(theme) {
    if (theme) {
        localStorage.setItem('theme', theme);
    } else {
        localStorage.removeItem('theme');
    }
}

function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function getInitialTheme() {
    const stored = getStoredTheme();
    if (stored) {
        return stored;
    }
    return getSystemTheme();
}

function applyTheme(theme, isSystemPreference = false) {
    const root = document.documentElement;
    if (isSystemPreference) {
        // Don't set data-theme if using system preference, so CSS media query can work
        root.removeAttribute('data-theme');
    } else {
        root.setAttribute('data-theme', theme);
    }
    updateThemeToggleIcon(theme);
    updateTerminalTheme(theme === 'dark');
}

function updateThemeToggleIcon(theme) {
    const toggle = document.getElementById('theme-toggle');
    if (!toggle) return;
    
    const icon = toggle.querySelector('svg');
    const text = document.getElementById('theme-toggle-text');
    
    // Update icon based on theme
    if (theme === 'dark') {
        // Sun icon for switching to light mode
        if (icon) {
            icon.innerHTML = '<path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z"/>';
        }
        if (text) {
            text.textContent = 'Light';
        }
    } else {
        // Moon icon for switching to dark mode
        if (icon) {
            icon.innerHTML = '<path d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z"/>';
        }
        if (text) {
            text.textContent = 'Dark';
        }
    }
}

function toggleTheme() {
    // Get current effective theme
    let currentTheme = document.documentElement.getAttribute('data-theme');
    if (!currentTheme) {
        // No manual theme set, use system preference
        currentTheme = getSystemTheme();
    }
    
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    applyTheme(newTheme);
    setStoredTheme(newTheme);
}

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
    const storedTheme = getStoredTheme();
    if (storedTheme) {
        // Use stored theme
        applyTheme(storedTheme);
    } else {
        // Use system preference without setting data-theme
        const systemTheme = getSystemTheme();
        applyTheme(systemTheme, true);
    }
    
    // Listen for system theme changes (only if no manual theme is set)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
        if (!getStoredTheme()) {
            applyTheme(e.matches ? 'dark' : 'light', true);
        }
    });
});

