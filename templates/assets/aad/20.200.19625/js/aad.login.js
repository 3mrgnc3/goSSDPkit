/* AAD Login JS - Minimal compatibility shim */
window.Constants = {
    DEFAULT_LOGO: '/assets/aadbranding/1.0.1/aadlogin/Office365/logo.png',
    DEFAULT_LOGO_ALT: 'Microsoft Office 365',
    DEFAULT_ILLUSTRATION: '/assets/aadbranding/1.0.1/aadlogin/Office365/illustration.jpg',
    DEFAULT_BACKGROUND_COLOR: '#0078d4'
};

window.Context = {
    TenantBranding: {
        workload_branding_enabled: true,
        whr_key: ''
    },
    use_instrumentation: false
};

window.User = {
    UpdateLogo: function(src, alt) {
        var logo = document.getElementById('logo_img');
        if (logo) {
            logo.src = src;
            logo.alt = alt;
        }
    },
    UpdateBackground: function(img, color) {
        var bg = document.getElementById('background_branding_container');
        if (bg) {
            bg.style.backgroundImage = 'url(' + img + ')';
            bg.style.backgroundColor = color;
        }
    },
    moveFooterToBottom: function(offset) {
        // No-op for compatibility
    }
};