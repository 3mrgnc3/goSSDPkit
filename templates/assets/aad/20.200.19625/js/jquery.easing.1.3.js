/* jQuery Easing 1.3 - Minimal compatibility */
window.jQuery = window.jQuery || window.$;
if (window.jQuery && window.jQuery.easing) {
    // Already loaded
} else if (window.jQuery) {
    window.jQuery.easing = {
        def: 'easeOutQuad',
        swing: function (x, t, b, c, d) {
            return window.jQuery.easing[window.jQuery.easing.def](x, t, b, c, d);
        },
        easeOutQuad: function (x, t, b, c, d) {
            return -c *(t/=d)*(t-2) + b;
        }
    };
}