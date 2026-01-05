/**
 * HTMX Authentication Interceptor
 * 
 * Automatically handles:
 * - Adding Authorization and CSRF headers to HTMX requests
 * - Detecting 401 responses and refreshing tokens
 * - Retrying requests with new tokens
 * - Redirecting to login on failed refresh
 */

// Intercept HTMX requests to add authentication headers
document.addEventListener('htmx:beforeRequest', function (event) {
    const accessToken = localStorage.getItem('access_token');
    const csrfToken = localStorage.getItem('csrf_token');
    console.log('[HTMX Auth] Preparing request with tokens');
    console.log("accessToken:", accessToken);
    console.log("csrfToken:", csrfToken);
    // Add Authorization header for all requests
    if (accessToken) {
        event.detail.headers = event.detail.headers || {};
        event.detail.headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Add CSRF token for state-changing requests
    const method = event.detail.verb || 'GET';
    if (csrfToken && ['POST', 'PUT', 'DELETE', 'PATCH'].includes(method.toUpperCase())) {
        event.detail.headers = event.detail.headers || {};
        event.detail.headers['X-CSRF-Token'] = csrfToken;
    }
});

// Handle HTMX response errors
document.addEventListener('htmx:responseError', async function (event) {
    const status = event.detail.xhr.status;

    // Only handle 401 Unauthorized
    if (status !== 401) {
        return;
    }

    console.log('[HTMX Auth] Token expired, attempting refresh...');

    const refreshToken = sessionStorage.getItem('refresh_token');
    if (!refreshToken) {
        console.error('[HTMX Auth] No refresh token available');
        window.location.href = '/login';
        return;
    }

    try {
        // Call refresh endpoint
        const response = await fetch('/api/auth/refresh', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': localStorage.getItem('csrf_token') || ''
            },
            body: JSON.stringify({
                refresh_token: refreshToken
            })
        });

        if (!response.ok) {
            throw new Error(`Refresh failed with status ${response.status}`);
        }

        const data = await response.json();

        // Update tokens
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('csrf_token', data.csrf_token);

        console.log('[HTMX Auth] Token refreshed successfully');
        console.log('[HTMX Auth] Retrying original request...');

        // Get the original request details
        const xhr = event.detail.xhr;
        const method = xhr.method || 'GET';
        const url = xhr.responseURL || event.detail.xhr.configUrl;

        // Prepare headers for retry
        const retryHeaders = {
            'Authorization': `Bearer ${data.access_token}`
        };

        if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(method.toUpperCase())) {
            retryHeaders['X-CSRF-Token'] = data.csrf_token;
        }

        // Get body if it exists
        const body = xhr.requestBody ? xhr.requestBody : null;

        // Retry with new tokens
        if (event.detail.target) {
            htmx.ajax(method, url, {
                target: event.detail.target,
                swap: event.detail.swap || 'outerHTML',
                headers: retryHeaders,
                ...(body && { values: body })
            });
        }

    } catch (error) {
        console.error('[HTMX Auth] Token refresh failed:', error);

        // Clear tokens and redirect to login
        localStorage.clear();
        sessionStorage.clear();
        window.location.href = '/login';
    }
});

// Optional: Add a global error handler for non-HTMX fetch calls
window.fetch_original = window.fetch;
window.fetch = async function (...args) {
    let response = await window.fetch_original(...args);

    // If 401 and not already retrying, refresh and retry
    if (response.status === 401 && !args[1]?._isRetry) {
        const refreshToken = sessionStorage.getItem('refresh_token');
        if (refreshToken) {
            try {
                const refreshResponse = await window.fetch_original('/api/auth/refresh', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': localStorage.getItem('csrf_token') || ''
                    },
                    body: JSON.stringify({ refresh_token: refreshToken })
                });

                if (refreshResponse.ok) {
                    const data = await refreshResponse.json();
                    localStorage.setItem('access_token', data.access_token);
                    localStorage.setItem('csrf_token', data.csrf_token);

                    // Retry with new token
                    const retryArgs = [...args];
                    retryArgs[1] = retryArgs[1] || {};
                    retryArgs[1].headers = retryArgs[1].headers || {};
                    retryArgs[1].headers['Authorization'] = `Bearer ${data.access_token}`;
                    retryArgs[1]._isRetry = true;

                    return window.fetch_original(...retryArgs);
                }
            } catch (e) {
                console.error('Token refresh error:', e);
            }
        }
    }

    return response;
};

// Show token expiration warning
function initTokenWarning() {
    function checkTokenExpiration() {
        const token = localStorage.getItem('access_token');
        if (!token) return;

        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            const expiresIn = (payload.exp * 1000) - Date.now();
            const warningTime = 2 * 60 * 1000; // 2 minutes

            if (expiresIn > 0 && expiresIn < warningTime) {
                console.warn('[HTMX Auth] Token expiring in', Math.floor(expiresIn / 1000), 'seconds');
            }
        } catch (e) {
            console.error('[HTMX Auth] Token parse error:', e);
        }
    }

    // Check every minute
    setInterval(checkTokenExpiration, 60000);
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTokenWarning);
} else {
    initTokenWarning();
}

console.log('[HTMX Auth] Interceptor loaded');
