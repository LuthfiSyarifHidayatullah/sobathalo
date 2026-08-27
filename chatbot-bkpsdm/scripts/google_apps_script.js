/**
 * ============================================================
 * Google Apps Script - Penerima Data Chatbot BKPSDM Bengkayang
 * ============================================================
 * 
 * CARA PENGGUNAAN:
 * 1. Buka Google Spreadsheet baru
 * 2. Buka menu Extensions > Apps Script
 * 3. Hapus kode default dan tempelkan seluruh kode ini
 * 4. Simpan project (beri nama misalnya "Chatbot BKPSDM Logger")
 * 5. Deploy sebagai Web App:
 *    - Klik Deploy > New deployment
 *    - Pilih type: Web app
 *    - Execute as: Me
 *    - Who has access: Anyone
 *    - Klik Deploy
 * 6. Salin URL Web App yang muncul ke file .env (GOOGLE_SCRIPT_URL)
 * 7. Set token rahasia di Script Properties (opsional):
 *    - Buka Project Settings (ikon gear)
 *    - Scroll ke Script Properties
 *    - Tambah property: Key = SECRET_TOKEN, Value = (token yang sama di .env)
 * 
 * STRUKTUR SPREADSHEET:
 * Sheet pertama akan digunakan dengan kolom:
 * A: Waktu
 * B: Pengguna
 * C: Bidang
 * D: Pelayanan
 * E: Info Diminta
 * F: Status
 */

// Fungsi utama untuk menerima POST request dari chatbot
function doPost(e) {
  try {
    // Parse data JSON dari request body
    var data = JSON.parse(e.postData.contents);

    // Validasi data minimal
    if (!data.waktu && !data.pengguna) {
      return ContentService.createTextOutput(
        JSON.stringify({ status: 'error', message: 'Data tidak lengkap' })
      ).setMimeType(ContentService.MimeType.JSON);
    }

    // Tulis ke spreadsheet
    writeToSheet(data);

    return ContentService.createTextOutput(
      JSON.stringify({ status: 'success', message: 'Data berhasil dicatat' })
    ).setMimeType(ContentService.MimeType.JSON);

  } catch (error) {
    return ContentService.createTextOutput(
      JSON.stringify({ status: 'error', message: error.toString() })
    ).setMimeType(ContentService.MimeType.JSON);
  }
}

// Fungsi untuk menulis data ke sheet
function writeToSheet(data) {
  var ss = SpreadsheetApp.getActiveSpreadsheet();
  var sheet = ss.getSheets()[0]; // Gunakan sheet pertama

  // Cek apakah header sudah ada, jika belum tambahkan
  if (sheet.getLastRow() === 0) {
    var headers = [
      'Waktu',
      'Pengguna',
      'Bidang',
      'Pelayanan',
      'Info Diminta',
      'Status'
    ];
    sheet.appendRow(headers);

    // Format header
    var headerRange = sheet.getRange(1, 1, 1, headers.length);
    headerRange.setFontWeight('bold');
    headerRange.setBackground('#4285f4');
    headerRange.setFontColor('#ffffff');

    // Freeze header row
    sheet.setFrozenRows(1);

    // Auto-resize kolom
    sheet.autoResizeColumns(1, headers.length);
  }

  // Tambahkan baris data baru
  var row = [
    data.waktu || '',
    data.pengguna || '',
    data.bidang || '',
    data.pelayanan || '',
    data.info_diminta || '',
    data.status || ''
  ];

  sheet.appendRow(row);

  // Pewarnaan baris berdasarkan status
  var lastRow = sheet.getLastRow();
  var rowRange = sheet.getRange(lastRow, 1, 1, 6);
  var status = (data.status || '').toLowerCase();

  if (status.indexOf('terjawab') !== -1) {
    rowRange.setBackground('#d4edda'); // Hijau muda
  } else if (status.indexOf('tidak dikenali') !== -1) {
    rowRange.setBackground('#f8d7da'); // Merah muda
  } else if (status.indexOf('dialihkan') !== -1 || status.indexOf('petugas') !== -1) {
    rowRange.setBackground('#fff3cd'); // Kuning muda
  }
}

// Fungsi GET untuk testing (buka URL di browser untuk cek)
function doGet(e) {
  return ContentService.createTextOutput(
    JSON.stringify({
      status: 'active',
      message: 'Chatbot BKPSDM Logger aktif',
      format: 'Waktu | Pengguna | Bidang | Pelayanan | Info Diminta | Status',
      timestamp: new Date().toISOString()
    })
  ).setMimeType(ContentService.MimeType.JSON);
}
