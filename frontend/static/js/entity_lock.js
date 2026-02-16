// Generic Entity Lock Management
// Automatically handles locking, unlocking, and refreshing locks for any entity

class EntityLockManager {
    constructor() {
        this.currentEntityId = null;
        this.currentEntityType = null;
        this.refreshInterval = null;
        this.activityCheckInterval = null;
        this.clientId = this.generateClientId();
        this.lastActivityTime = Date.now();
        this.REFRESH_INTERVAL_MS = 5 * 60 * 1000; // Check every 5 minutes
        this.INACTIVITY_TIMEOUT_MS = 20 * 60 * 1000; // 20 minutes of inactivity
        this.activityEvents = ['mousedown', 'keydown', 'scroll', 'touchstart'];
    }

    // Generate a unique client ID for this browser tab
    generateClientId() {
        // Try to get existing clientId from sessionStorage (survives page refreshes within same tab)
        let clientId = sessionStorage.getItem('entity_client_id');
        if (!clientId) {
            // Generate new unique ID using crypto API or fallback to timestamp
            if (window.crypto && window.crypto.randomUUID) {
                clientId = crypto.randomUUID();
            } else {
                clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            }
            sessionStorage.setItem('entity_client_id', clientId);
        }
        return clientId;
    }

    // Get headers with client ID
    getHeaders() {
        return {
            'Content-Type': 'application/json',
            'X-Client-ID': this.clientId
        };
    }

    // Track user activity
    trackActivity() {
        this.lastActivityTime = Date.now();
    }

    // Start managing an entity lock
    // entityType: e.g., 'nalozi', 'partneri', 'users'
    // entityId: the ID of the entity
    // options: { refreshInterval, inactivityTimeout, entityName, unlockUrl, refreshUrl }
    startLock(entityType, entityId, options = {}) {
        this.currentEntityType = entityType;
        this.currentEntityId = entityId;
        this.lastActivityTime = Date.now();
        this.options = {
            refreshInterval: options.refreshInterval || this.REFRESH_INTERVAL_MS,
            inactivityTimeout: options.inactivityTimeout || this.INACTIVITY_TIMEOUT_MS,
            entityName: options.entityName || entityType,
            unlockUrl: options.unlockUrl || `/api/${entityType}/unlock/${entityId}`,
            refreshUrl: options.refreshUrl || `/api/${entityType}/refresh-lock/${entityId}`
        };
        
        // Setup activity tracking
        this.activityHandler = () => this.trackActivity();
        this.activityEvents.forEach(event => {
            document.addEventListener(event, this.activityHandler, true);
        });
        
        // Check activity and refresh lock periodically
        this.activityCheckInterval = setInterval(() => {
            this.checkActivityAndRefresh();
        }, this.options.refreshInterval);

        // Unlock on page unload
        this.unloadHandler = () => this.unlockOnUnload();
        window.addEventListener('beforeunload', this.unloadHandler);
        
        console.log(`Lock manager started for ${entityType} ${entityId} with clientId ${this.clientId}`);
        console.log(`Inactivity timeout: ${this.options.inactivityTimeout / 60000} minutes`);
    }

    // Check activity and handle inactivity
    checkActivityAndRefresh() {
        const inactiveTime = Date.now() - this.lastActivityTime;
        const remainingTime = this.options.inactivityTimeout - inactiveTime;
        
        // Warn user 2 minutes before timeout
        if (remainingTime > 0 && remainingTime <= 2 * 60 * 1000 && remainingTime > this.options.refreshInterval - 1000) {
            const minutesLeft = Math.ceil(remainingTime / 60000);
            console.warn(`Inactivity warning: ${minutesLeft} minute(s) remaining`);
            this.showInactivityWarning(minutesLeft);
        }
        
        // If inactive for too long, auto-close and unlock
        if (inactiveTime >= this.options.inactivityTimeout) {
            console.warn('Inactivity timeout reached, closing...');
            this.handleInactivityTimeout();
            return;
        }
        
        // Refresh lock (server-side timeout prevention)
        this.refreshLock();
    }

    // Handle inactivity timeout
    handleInactivityTimeout() {
        alert(`Sesija za uređivanje ${this.options.entityName} je istekla zbog neaktivnosti. Dijalog će biti zatvoren.`);
        this.unlock().then(() => {
            // Close dialog/form
            window.location.reload(); // Or use htmx to navigate back
        });
    }

    // Show warning (can be customized with better UI)
    showInactivityWarning(minutesLeft) {
        // You can implement a nicer toast/notification here
        const notification = document.getElementById('notification');
        if (notification) {
            notification.textContent = `Upozorenje: Sesija ističe za ${minutesLeft} minut(a) zbog neaktivnosti`;
            notification.className = 'fixed top-5 right-5 px-4 py-3 bg-yellow-500 text-white rounded';
            setTimeout(() => {
                notification.className = 'fixed top-5 right-5 px-4 py-3 bg-green-500 text-white rounded hidden';
            }, 5000);
        }
    }

    // Refresh the lock to prevent timeout
    async refreshLock() {
        if (!this.currentEntityId || !this.currentEntityType) return;

        try {
            const response = await fetch(this.options.refreshUrl, {
                method: 'POST',
                headers: this.getHeaders()
            });

            const data = await response.json();
            if (!data.success) {
                console.warn('Failed to refresh lock:', data.message);
                this.stopLock();
                // Optionally notify user that their lock expired
                alert(`Vaša sesija za uređivanje ${this.options.entityName} je istekla. Molimo zatvorite i ponovo otvorite.`);
            }
        } catch (error) {
            console.error('Error refreshing lock:', error);
        }
    }

    // Unlock the entity
    async unlock() {
        if (!this.currentEntityId || !this.currentEntityType) return Promise.resolve();

        const unlockUrl = this.options.unlockUrl;
        this.stopLock();

        try {
            // Use regular fetch for explicit unlock (not during page unload)
            const response = await fetch(unlockUrl, {
                method: 'POST',
                headers: this.getHeaders()
            });
            
            if (!response.ok) {
                console.error(`Failed to unlock ${this.currentEntityType}:`, response.status);
            }
            return response;
        } catch (error) {
            console.error(`Error unlocking ${this.currentEntityType}:`, error);
            throw error;
        }
    }
    
    // Unlock during page unload (fire and forget)
    unlockOnUnload() {
        if (!this.currentEntityId || !this.currentEntityType) return;
        
        try {
            // Try both approaches for reliability
            const url = `${this.options.unlockUrl}?client_id=${encodeURIComponent(this.clientId)}`;
            
            // Method 1: sendBeacon (most reliable for page unload)
            const blob = new Blob([JSON.stringify({})], { type: 'application/json' });
            const beaconSent = navigator.sendBeacon(url, blob);
            
            // Method 2: Synchronous XMLHttpRequest as fallback (deprecated but reliable)
            if (!beaconSent) {
                const xhr = new XMLHttpRequest();
                xhr.open('POST', url, false); // false = synchronous
                xhr.setRequestHeader('X-Client-ID', this.clientId);
                xhr.send();
            }
        } catch (error) {
            console.error(`Error unlocking ${this.currentEntityType} on unload:`, error);
        }
    }

    // Stop managing the lock
    stopLock() {
        if (this.activityCheckInterval) {
            clearInterval(this.activityCheckInterval);
            this.activityCheckInterval = null;
        }
        if (this.activityHandler) {
            this.activityEvents.forEach(event => {
                document.removeEventListener(event, this.activityHandler, true);
            });
            this.activityHandler = null;
        }
        if (this.unloadHandler) {
            window.removeEventListener('beforeunload', this.unloadHandler);
            this.unloadHandler = null;
        }
        this.currentEntityId = null;
        this.currentEntityType = null;
        this.options = null;
    }
}

// Global instance
const entityLockManager = new EntityLockManager();

// On page load, cleanup any stale locks from previous session (e.g., after browser refresh)
window.addEventListener('DOMContentLoaded', function() {
    // Release any locks that might be stuck from previous page session
    fetch('/api/locks/cleanup', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Client-ID': entityLockManager.clientId
        }
    }).catch(err => console.log('Lock cleanup failed:', err));
});

// Helper function to be called when opening an entity for editing
// Usage: startEntityLock('nalozi', 123, { entityName: 'naloga' })
function startEntityLock(entityType, entityId, options = {}) {
    entityLockManager.startLock(entityType, entityId, options);
}

// Helper function to be called when closing/saving an entity
function unlockEntity() {
    return entityLockManager.unlock();
}

// Add client ID to all HTMX requests
document.body.addEventListener('htmx:configRequest', function(event) {
    event.detail.headers['X-Client-ID'] = entityLockManager.clientId;
});

// Handle close button with unlock (generic function)
async function handleCloseWithUnlock(event, navigateUrl, targetElement = '#content') {
    event.preventDefault();
    event.stopPropagation();
    
    // Unlock first (wait for it to complete)
    await entityLockManager.unlock();
    
    // Then navigate
    htmx.ajax('GET', navigateUrl, {target: targetElement});
}
