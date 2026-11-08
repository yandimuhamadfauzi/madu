document.addEventListener("DOMContentLoaded", function() {
    var mapElement = document.getElementById('map');
    if (!mapElement) return;

    var bounds = [[-90, -180], [90, 180]];

    var map = L.map('map', {
        minZoom: 2,
        maxBounds: bounds,
        maxBoundsViscosity: 1.0
    }).setView([20, 0], 2);
    
    L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Base/MapServer/tile/{z}/{y}/{x}', {
        maxZoom: 16,
        noWrap: true,
        attribution: 'Tiles © Esri — Esri, DeLorme, NAVTEQ'
    }).addTo(map);

    setTimeout(function(){ map.invalidateSize(); }, 500);

    // Inisialisasi cluster group
    var markers = L.markerClusterGroup({
        showCoverageOnHover: false,
        maxClusterRadius: 40
    });

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
                    var marker = L.marker([data.latitude, data.longitude]);
                    var isp = data.organization_name || data.organization || "Tidak diketahui";
                    
                    marker.bindPopup(
                        "<b>IP:</b> " + ip + 
                        "<br><b>Lokasi:</b> " + (data.city ? data.city + ", " : "") + data.country + 
                        "<br><b>ISP:</b> " + isp
                    );

                    markers.addLayer(marker);
                }
            })
            .catch(err => console.error("Gagal melacak lokasi IP: ", ip));
        });

        map.addLayer(markers);
    } catch (error) {
        console.error("Gagal memproses data IP:", error);
    }
});