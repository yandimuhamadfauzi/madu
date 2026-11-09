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

document.addEventListener('DOMContentLoaded', function() {
    // Cari semua elemen yang memiliki class 'btn-cek-ip'
    const ipLinks = document.querySelectorAll('.btn-cek-ip');
    
    // Pasang 'pendengar kejadian' (event listener) pada masing-masing link
    ipLinks.forEach(function(link) {
        link.addEventListener('click', async function(event) {
            event.preventDefault(); // Cegah browser melompat ke atas halaman
            
            const ip = this.getAttribute('data-ip');
            
            try {
                // Panggil API internal Madu
                const response = await fetch(`/api/ip-info?ip=${ip}`);
                
                if (!response.ok) {
                    throw new Error(`Gagal menghubungi server. Status HTTP: ${response.status}`);
                }
                
                const data = await response.json();
                
                // Format hasil
                let pesan = `🔎 HASIL INVESTIGASI IP: ${data.ipAddress}\n\n`;
                pesan += `⚠️ Skor Bahaya: ${data.abuseConfidenceScore}%\n`;
                pesan += `🌍 Negara: ${data.countryCode}\n`;
                pesan += `🏢 ISP: ${data.isp}\n`;
                pesan += `📡 Tipe: ${data.usageType}\n`;
                pesan += `🚨 Total Laporan: ${data.totalReports} kali`;

                alert(pesan);

            } catch (error) {
                alert("Terjadi kesalahan saat mengecek IP:\n" + error.message);
            }
        });
    });
});