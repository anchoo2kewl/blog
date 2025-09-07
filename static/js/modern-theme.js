// Modern Theme JavaScript
// Enhanced functionality for the modern blog theme

(function() {
    'use strict';
    
    // Theme utilities
    const Theme = {
        // Initialize theme
        init() {
            this.setupThemeToggle();
            this.setupMobileMenu();
            this.setupSearch();
            this.setupScrollEffects();
            this.setupAnimations();
            this.setupTooltips();
            this.setupCodeBlocks();
            this.setupImageLightbox();
        },
        
        // Theme toggle functionality
        setupThemeToggle() {
            const themeToggle = document.getElementById('theme-toggle');
            const html = document.documentElement;
            
            if (!themeToggle) return;
            
            // Set initial theme based on localStorage or system preference
            const savedTheme = localStorage.getItem('theme') || 
                              (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
            
            if (savedTheme === 'dark') {
                html.classList.add('dark');
            }
            
            themeToggle.addEventListener('click', () => {
                const isDark = html.classList.contains('dark');
                
                if (isDark) {
                    html.classList.remove('dark');
                    localStorage.setItem('theme', 'light');
                } else {
                    html.classList.add('dark');
                    localStorage.setItem('theme', 'dark');
                }
                
                // Animate the transition
                document.body.style.transition = 'background-color 0.3s ease, color 0.3s ease';
                setTimeout(() => {
                    document.body.style.transition = '';
                }, 300);
            });
            
            // Listen for system theme changes
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
                if (!localStorage.getItem('theme')) {
                    if (e.matches) {
                        html.classList.add('dark');
                    } else {
                        html.classList.remove('dark');
                    }
                }
            });
        },
        
        // Mobile menu functionality
        setupMobileMenu() {
            const mobileMenuBtn = document.getElementById('mobile-menu-btn');
            const mobileMenu = document.getElementById('mobile-menu');
            
            if (!mobileMenuBtn || !mobileMenu) return;
            
            mobileMenuBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                mobileMenu.classList.toggle('open');
                
                // Update button icon
                const icon = mobileMenuBtn.querySelector('svg');
                if (mobileMenu.classList.contains('open')) {
                    icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>';
                } else {
                    icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>';
                }
            });
            
            // Close mobile menu when clicking outside
            document.addEventListener('click', (e) => {
                if (!mobileMenuBtn.contains(e.target) && !mobileMenu.contains(e.target)) {
                    mobileMenu.classList.remove('open');
                    const icon = mobileMenuBtn.querySelector('svg');
                    icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>';
                }
            });
            
            // Close mobile menu on escape key
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && mobileMenu.classList.contains('open')) {
                    mobileMenu.classList.remove('open');
                    const icon = mobileMenuBtn.querySelector('svg');
                    icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>';
                }
            });
        },
        
        // Search functionality
        setupSearch() {
            const searchToggle = document.getElementById('search-toggle');
            
            if (!searchToggle) return;
            
            searchToggle.addEventListener('click', () => {
                this.showSearchModal();
            });
            
            // Keyboard shortcut for search (Ctrl/Cmd + K)
            document.addEventListener('keydown', (e) => {
                if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                    e.preventDefault();
                    this.showSearchModal();
                }
            });
        },
        
        // Show search modal
        showSearchModal() {
            const modal = this.createSearchModal();
            document.body.appendChild(modal);
            
            // Focus on input
            const input = modal.querySelector('input');
            setTimeout(() => input.focus(), 100);
            
            // Setup search functionality
            this.setupSearchInput(input, modal);
        },
        
        // Create search modal
        createSearchModal() {
            const modal = document.createElement('div');
            modal.className = 'fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-start justify-center pt-20';
            modal.innerHTML = `
                <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl mx-4 overflow-hidden">
                    <div class="flex items-center p-4 border-b border-gray-200 dark:border-gray-700">
                        <svg class="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                        </svg>
                        <input 
                            type="text" 
                            placeholder="Search posts..." 
                            class="flex-1 bg-transparent outline-none text-gray-900 dark:text-gray-100"
                        />
                        <button class="ml-3 text-gray-400 hover:text-gray-600">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                            </svg>
                        </button>
                    </div>
                    <div class="max-h-80 overflow-y-auto p-4" id="search-results">
                        <div class="text-center py-8 text-gray-500">
                            <svg class="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                            </svg>
                            <p>Start typing to search posts...</p>
                        </div>
                    </div>
                </div>
            `;
            
            // Close modal events
            const closeBtn = modal.querySelector('button');
            closeBtn.addEventListener('click', () => modal.remove());
            
            modal.addEventListener('click', (e) => {
                if (e.target === modal) modal.remove();
            });
            
            document.addEventListener('keydown', function escHandler(e) {
                if (e.key === 'Escape') {
                    modal.remove();
                    document.removeEventListener('keydown', escHandler);
                }
            });
            
            return modal;
        },
        
        // Setup search input functionality
        setupSearchInput(input, modal) {
            let searchTimeout;
            
            input.addEventListener('input', (e) => {
                const query = e.target.value.trim();
                
                clearTimeout(searchTimeout);
                
                if (query.length < 2) {
                    this.showSearchPlaceholder();
                    return;
                }
                
                searchTimeout = setTimeout(() => {
                    this.performSearch(query);
                }, 300);
            });
        },
        
        // Perform search
        async performSearch(query) {
            const resultsContainer = document.getElementById('search-results');
            
            // Show loading
            resultsContainer.innerHTML = `
                <div class="text-center py-8">
                    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p class="text-gray-500">Searching...</p>
                </div>
            `;
            
            try {
                // This would be replaced with actual search API call
                const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
                const results = await response.json();
                
                this.displaySearchResults(results);
            } catch (error) {
                console.error('Search error:', error);
                this.showSearchError();
            }
        },
        
        // Display search results
        displaySearchResults(results) {
            const resultsContainer = document.getElementById('search-results');
            
            if (results.length === 0) {
                resultsContainer.innerHTML = `
                    <div class="text-center py-8 text-gray-500">
                        <svg class="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                        </svg>
                        <p>No posts found for your search.</p>
                    </div>
                `;
                return;
            }
            
            const resultsHTML = results.map(result => `
                <a href="/blog/${result.slug}" class="block p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                    <h3 class="font-semibold mb-1">${result.title}</h3>
                    <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">${result.excerpt}</p>
                    <div class="text-xs text-gray-500">${result.date}</div>
                </a>
            `).join('');
            
            resultsContainer.innerHTML = resultsHTML;
        },
        
        // Show search placeholder
        showSearchPlaceholder() {
            const resultsContainer = document.getElementById('search-results');
            if (resultsContainer) {
                resultsContainer.innerHTML = `
                    <div class="text-center py-8 text-gray-500">
                        <svg class="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                        </svg>
                        <p>Start typing to search posts...</p>
                    </div>
                `;
            }
        },
        
        // Show search error
        showSearchError() {
            const resultsContainer = document.getElementById('search-results');
            if (resultsContainer) {
                resultsContainer.innerHTML = `
                    <div class="text-center py-8 text-red-500">
                        <svg class="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                        <p>Search is temporarily unavailable. Please try again later.</p>
                    </div>
                `;
            }
        },
        
        // Setup scroll effects
        setupScrollEffects() {
            // Back to top button
            const backToTopBtn = document.getElementById('back-to-top');
            
            if (backToTopBtn) {
                window.addEventListener('scroll', () => {
                    if (window.scrollY > 300) {
                        backToTopBtn.classList.remove('opacity-0', 'invisible');
                        backToTopBtn.classList.add('opacity-100', 'visible');
                    } else {
                        backToTopBtn.classList.add('opacity-0', 'invisible');
                        backToTopBtn.classList.remove('opacity-100', 'visible');
                    }
                });
                
                backToTopBtn.addEventListener('click', () => {
                    window.scrollTo({ top: 0, behavior: 'smooth' });
                });
            }
            
            // Navbar scroll effect
            const navbar = document.querySelector('.navbar');
            if (navbar) {
                let lastScrollY = window.scrollY;
                
                window.addEventListener('scroll', () => {
                    const currentScrollY = window.scrollY;
                    
                    if (currentScrollY > 100) {
                        navbar.classList.add('navbar-scrolled');
                    } else {
                        navbar.classList.remove('navbar-scrolled');
                    }
                    
                    lastScrollY = currentScrollY;
                });
            }
        },
        
        // Setup animations
        setupAnimations() {
            // Intersection Observer for fade-in animations
            const observerOptions = {
                threshold: 0.1,
                rootMargin: '0px 0px -50px 0px'
            };
            
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        entry.target.classList.add('animate-fade-in');
                        observer.unobserve(entry.target);
                    }
                });
            }, observerOptions);
            
            // Observe elements for animation
            const animateElements = document.querySelectorAll('.blog-post, .hero, .admin-dashboard > *');
            animateElements.forEach(el => {
                observer.observe(el);
            });
        },
        
        // Setup tooltips
        setupTooltips() {
            const tooltipElements = document.querySelectorAll('[data-tooltip]');
            
            tooltipElements.forEach(element => {
                element.addEventListener('mouseenter', (e) => {
                    this.showTooltip(e.target, e.target.dataset.tooltip);
                });
                
                element.addEventListener('mouseleave', () => {
                    this.hideTooltip();
                });
            });
        },
        
        // Show tooltip
        showTooltip(element, text) {
            const tooltip = document.createElement('div');
            tooltip.className = 'tooltip absolute z-50 px-2 py-1 text-sm bg-gray-900 text-white rounded shadow-lg';
            tooltip.textContent = text;
            tooltip.id = 'tooltip';
            
            document.body.appendChild(tooltip);
            
            const rect = element.getBoundingClientRect();
            tooltip.style.top = `${rect.top - tooltip.offsetHeight - 5}px`;
            tooltip.style.left = `${rect.left + (rect.width - tooltip.offsetWidth) / 2}px`;
        },
        
        // Hide tooltip
        hideTooltip() {
            const tooltip = document.getElementById('tooltip');
            if (tooltip) {
                tooltip.remove();
            }
        },
        
        // Setup code blocks
        setupCodeBlocks() {
            const codeBlocks = document.querySelectorAll('pre code');
            
            codeBlocks.forEach(block => {
                const pre = block.parentElement;
                pre.classList.add('relative');
                
                // Add copy button
                const copyBtn = document.createElement('button');
                copyBtn.className = 'absolute top-2 right-2 px-2 py-1 text-xs bg-gray-700 text-white rounded opacity-0 hover:opacity-100 transition-opacity';
                copyBtn.textContent = 'Copy';
                
                copyBtn.addEventListener('click', () => {
                    navigator.clipboard.writeText(block.textContent);
                    copyBtn.textContent = 'Copied!';
                    setTimeout(() => {
                        copyBtn.textContent = 'Copy';
                    }, 2000);
                });
                
                pre.appendChild(copyBtn);
                
                // Show copy button on hover
                pre.addEventListener('mouseenter', () => {
                    copyBtn.classList.add('opacity-100');
                });
                
                pre.addEventListener('mouseleave', () => {
                    copyBtn.classList.remove('opacity-100');
                });
            });
        },
        
        // Setup image lightbox
        setupImageLightbox() {
            const images = document.querySelectorAll('article img');
            
            images.forEach(img => {
                img.style.cursor = 'pointer';
                img.addEventListener('click', () => {
                    this.showLightbox(img.src, img.alt);
                });
            });
        },
        
        // Show image lightbox
        showLightbox(src, alt) {
            const lightbox = document.createElement('div');
            lightbox.className = 'fixed inset-0 bg-black/90 z-50 flex items-center justify-center p-4';
            lightbox.innerHTML = `
                <div class="relative max-w-full max-h-full">
                    <img src="${src}" alt="${alt}" class="max-w-full max-h-full object-contain rounded-lg">
                    <button class="absolute top-4 right-4 text-white hover:text-gray-300 text-2xl">×</button>
                </div>
            `;
            
            document.body.appendChild(lightbox);
            
            // Close lightbox
            const closeBtn = lightbox.querySelector('button');
            closeBtn.addEventListener('click', () => lightbox.remove());
            
            lightbox.addEventListener('click', (e) => {
                if (e.target === lightbox) lightbox.remove();
            });
            
            document.addEventListener('keydown', function escHandler(e) {
                if (e.key === 'Escape') {
                    lightbox.remove();
                    document.removeEventListener('keydown', escHandler);
                }
            });
        }
    };
    
    // Initialize theme when DOM is loaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => Theme.init());
    } else {
        Theme.init();
    }
    
    // Update year in footer
    const yearElement = document.getElementById('current-year');
    if (yearElement) {
        yearElement.textContent = new Date().getFullYear();
    }
    
    // Add navbar scrolled effect styles
    const style = document.createElement('style');
    style.textContent = `
        .navbar-scrolled {
            backdrop-filter: blur(20px);
            background: rgba(255, 255, 255, 0.9);
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        
        .dark .navbar-scrolled {
            background: rgba(3, 7, 18, 0.9);
            box-shadow: 0 1px 3px rgba(255, 255, 255, 0.1);
        }
        
        .animate-fade-in {
            animation: fade-in 0.6s ease-out forwards;
        }
        
        @keyframes fade-in {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
    `;
    document.head.appendChild(style);
    
    // Make Theme available globally for debugging
    window.Theme = Theme;
})();