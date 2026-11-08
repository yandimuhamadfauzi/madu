document.addEventListener("DOMContentLoaded", function() {
    const canvas = document.getElementById('trendChart');
    if (!canvas) return;

    let labels = [];
    let dataPoin = [];

    try {
        labels = JSON.parse(canvas.getAttribute('data-labels') || '[]');
        dataPoin = JSON.parse(canvas.getAttribute('data-counts') || '[]');
    } catch (e) {
        console.error("Gagal membaca data tren serangan:", e);
    }

    new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Total Request Berbahaya',
                data: dataPoin,
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.2)',
                borderWidth: 2,
                fill: true,
                tension: 0.4,
                pointRadius: 3,
                pointBackgroundColor: '#3b82f6'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { labels: { color: '#a0aec0' } }
            },
            scales: {
                x: { 
                    ticks: { color: '#a0aec0' },
                    grid: { color: 'rgba(160, 174, 192, 0.1)' }
                },
                y: { 
                    beginAtZero: true,
                    ticks: { color: '#a0aec0', precision: 0 },
                    grid: { color: 'rgba(160, 174, 192, 0.1)' }
                }
            }
        }
    });
});