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
 * 7. Set token rahasia di Script Properties:
 *    - Buka Project Settings (ikon gear)
 *    - Scroll ke Script Properties
 *    - Tambah property: Key = SECRET_TOKEN, Value = (token yang sama di .env)
 * 
 * STRUKTUR SPREADSHEET:
 * Sheet pertama akan digunakan dengan kolom:
 * A: Timestamp
 * B: Message ID
 * C: User ID (Hash)
 * D: Nama WhatsApp
 * E: Bidang
 * F: Pelayanan
 * G: Jenis Informasi
 * H: Pesan Asli
 * I: Status
 */

// Fungsi utama untuk menerima POST request dari chatbot
function doPost(e) {
  try {
    // Verifikasi token (opsional tapi direkomendasikan)
    var secretToken = PropertiesService.getScriptProperties().getProperty('SECRET_TOKEN');
    if (secretToken) {
      var authHeader = e.parameter.token || '';
      // Cek dari header Authorization
      if (e.postData && e.postData.headers) {
        var headers = e.postData.headers;
        if (headers['Authorization']) {
          authHeader = headers['Authorization'].replace('Bearer ', '');
        }
      }
      // Alternatif: cek dari parameter
      if (!authHeader && e.parameter && e.parameter.token) {
        authHeader = e.parameter.token;
      }
    }

    // Parse data JSON dari request body
    var data = JSON.parse(e.postData.contents);

    // Validasi data
    if (!data.timestamp || !data.message_id) {
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
      'Timestamp',
      'Message ID', 
      'User ID (Hash)',
      'Nama WhatsApp',
      'Bidang',
      'Pelayanan',
      'Jenis Informasi',
      'Pesan Asli',
      'Status'
    ];
    sheet.appendRow(headers);

    // Format header
    var headerRange = sheet.getRange(1, 1, 1, headers.length);
    headerRange.setFontWeight('bold');
    headerRange.setBackground('#4285f4');
    headerRange.setFontColor('#ffffff');
  }

  // Tambahkan baris data baru
  var row = [
    data.timestamp || '',
    data.message_id || '',
    data.user_id_hash || '',
    data.push_name || '',
    data.bidang || '',
    data.pelayanan || '',
    data.jenis_info || '',
    data.pesan_asli || '',
    data.status || ''
  ];

  sheet.appendRow(row);

  // Pewarnaan baris berdasarkan status
  var lastRow = sheet.getLastRow();
  var statusCell = sheet.getRange(lastRow, 9);
  var status = data.status || '';

  if (status === 'terjawab') {
    statusCell.setBackground('#d4edda'); // Hijau muda
  } else if (status === 'tidak_dikenali') {
    statusCell.setBackground('#f8d7da'); // Merah muda
  } else if (status === 'dialihkan_ke_petugas') {
    statusCell.setBackground('#fff3cd'); // Kuning muda
  }
}

// Fungsi GET untuk testing (buka URL di browser untuk cek)
function doGet(e) {
  return ContentService.createTextOutput(
    JSON.stringify({
      status: 'active',
      message: 'Chatbot BKPSDM Logger aktif',
      timestamp: new Date().toISOString()
    })
  ).setMimeType(ContentService.MimeType.JSON);
}

// Fungsi untuk membuat sheet ringkasan statistik (opsional)
function createSummary() {
  var ss = SpreadsheetApp.getActiveSpreadsheet();
  var dataSheet = ss.getSheets()[0];
  
  // Cek apakah sheet "Ringkasan" sudah ada
  var summarySheet = ss.getSheetByName('Ringkasan');
  if (!summarySheet) {
    summarySheet = ss.insertSheet('Ringkasan');
  } else {
    summarySheet.clear();
  }

  var data = dataSheet.getDataRange().getValues();
  if (data.length <= 1) return; // Hanya header

  // Hitung statistik
  var stats = {
    total: data.length - 1,
    terjawab: 0,
    tidak_dikenali: 0,
    dialihkan: 0,
    perBidang: {},
    perPelayanan: {}
  };

  for (var i = 1; i < data.length; i++) {
    var status = data[i][8];
    var bidang = data[i][4];
    var pelayanan = data[i][5];

    if (status === 'terjawab') stats.terjawab++;
    else if (status === 'tidak_dikenali') stats.tidak_dikenali++;
    else if (status === 'dialihkan_ke_petugas') stats.dialihkan++;

    if (bidang) {
      stats.perBidang[bidang] = (stats.perBidang[bidang] || 0) + 1;
    }
    if (pelayanan) {
      stats.perPelayanan[pelayanan] = (stats.perPelayanan[pelayanan] || 0) + 1;
    }
  }

  // Tulis ringkasan
  summarySheet.getRange('A1').setValue('RINGKASAN STATISTIK CHATBOT BKPSDM');
  summarySheet.getRange('A1').setFontWeight('bold').setFontSize(14);
  
  summarySheet.getRange('A3').setValue('Total Pesan:');
  summarySheet.getRange('B3').setValue(stats.total);
  
  summarySheet.getRange('A4').setValue('Terjawab:');
  summarySheet.getRange('B4').setValue(stats.terjawab);
  
  summarySheet.getRange('A5').setValue('Tidak Dikenali:');
  summarySheet.getRange('B5').setValue(stats.tidak_dikenali);
  
  summarySheet.getRange('A6').setValue('Dialihkan ke Petugas:');
  summarySheet.getRange('B6').setValue(stats.dialihkan);

  // Statistik per bidang
  summarySheet.getRange('A8').setValue('STATISTIK PER BIDANG');
  summarySheet.getRange('A8').setFontWeight('bold');
  var row = 9;
  for (var bidang in stats.perBidang) {
    summarySheet.getRange('A' + row).setValue(bidang);
    summarySheet.getRange('B' + row).setValue(stats.perBidang[bidang]);
    row++;
  }

  // Statistik per pelayanan
  row += 1;
  summarySheet.getRange('A' + row).setValue('STATISTIK PER PELAYANAN');
  summarySheet.getRange('A' + row).setFontWeight('bold');
  row++;
  for (var pelayanan in stats.perPelayanan) {
    summarySheet.getRange('A' + row).setValue(pelayanan);
    summarySheet.getRange('B' + row).setValue(stats.perPelayanan[pelayanan]);
    row++;
  }

  // Auto-resize kolom
  summarySheet.autoResizeColumns(1, 2);
}
