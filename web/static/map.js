document.addEventListener("DOMContentLoaded", function() {
    var mapElement = document.getElementById('map');
    if (!mapElement) return;

    // Batasi area peta agar tidak bisa digeser jauh keluar dunia
    var bounds = [[-90, -180], [90, 180]];

    var map = L.map('map', {
        minZoom: 2,             // Mencegah user terlalu zoom out
        maxBounds: bounds,      // Mengunci area agar tidak keluar peta
        maxBoundsViscosity: 1.0 // Kekuatan pantulan saat mencapai batas
    }).setView([20, 0], 2);
    
    L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Base/MapServer/tile/{z}/{y}/{x}', {
        maxZoom: 16,
        noWrap: true,           // Mencegah peta berulang menyamping saat di-scroll
        attribution: 'Tiles © Esri — Esri, DeLorme, NAVTEQ'
    }).addTo(map);

    setTimeout(function(){ map.invalidateSize(); }, 500);

    try {
        var rawData = mapElement.getAttribute('data-ips');
        if (!rawData) return;
        
        var ipList = JSON.parse(rawData);
        var uniqueIPs = [...new Set(ipList)].filter(ip => ip.trim() !== "");

        uniqueIPs.forEach(function(ip) {
            fetch('https://get.geojs.io/v1/ip/geo/' + ip + '.json')
            .then(response => response.json())
            .then(data => {
                if(data.latitude && data.longitude) {
                    var marker = L.marker([data.latitude, data.longitude]).addTo(map);
                    
                    // Deteksi ISP dari GeoJS (menggunakan organization_name atau organization)
                    var isp = data.organization_name || data.organization || "Tidak diketahui";
                    
                    // Tambahkan ISP ke dalam tampilan popup
                    marker.bindPopup(
                        "<b>IP:</b> " + ip + 
                        "<br><b>Lokasi:</b> " + (data.city ? data.city + ", " : "") + data.country + 
                        "<br><b>ISP:</b> " + isp
                    );
                }
            })
            .catch(err => console.error("Gagal melacak lokasi IP: ", ip));
        });
    } catch (error) {
        console.error("Gagal memproses data IP:", error);
    }
});