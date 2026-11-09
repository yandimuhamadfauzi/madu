document.addEventListener("DOMContentLoaded", function() {
    function setupToggle(toggleId, inputId, openId, closedId) {
        const toggleBtn = document.getElementById(toggleId);
        if (!toggleBtn) return;

        toggleBtn.addEventListener("click", function() {
            const input = document.getElementById(inputId);
            const eyeOpen = document.getElementById(openId);
            const eyeClosed = document.getElementById(closedId);

            if (!input || !eyeOpen || !eyeClosed) return;

            if (input.type === "password") {
                input.type = "text";
                eyeOpen.style.display = "none";
                eyeClosed.style.display = "block";
            } else {
                input.type = "password";
                eyeOpen.style.display = "block";
                eyeClosed.style.display = "none";
            }
        });
    }

    // Eksekusi fungsi untuk elemen-elemen form
    setupToggle("toggle-p", "p", "eye-open-p", "eye-closed-p");
    setupToggle("toggle-p2", "p2", "eye-open-p2", "eye-closed-p2");
});