#include <SPI.h>
#include <MFRC522.h>
#include <Wire.h>
#include <LiquidCrystal_I2C.h>

#define SS_PIN 10
#define RST_PIN 9

#define LED_R 6
#define LED_G 3
#define LED_B 5

#define LCD_ADDR 0x27

MFRC522 mfrc522(SS_PIN, RST_PIN);
LiquidCrystal_I2C lcd(LCD_ADDR, 16, 2);

const unsigned long PYTHON_TIMEOUT_MS = 12000;
const unsigned long SCAN_COOLDOWN_MS  = 1500;
unsigned long lastScanMs = 0;

void setColor(bool r, bool g, bool b) {
  digitalWrite(LED_R, r);
  digitalWrite(LED_G, g);
  digitalWrite(LED_B, b);
}

void blinkColor(bool r, bool g, bool b, int times, int onMs = 150, int offMs = 150) {
  for (int i = 0; i < times; i++) {
    setColor(r, g, b);
    delay(onMs);
    setColor(0, 0, 0);
    delay(offMs);
  }
}

void lcdShow(const String& l1, const String& l2) {
  lcd.clear();
  lcd.setCursor(0, 0);
  lcd.print(l1.substring(0, 16));
  lcd.setCursor(0, 1);
  lcd.print(l2.substring(0, 16));
}

String uidToString(MFRC522::Uid *uid) {
  String s = "";
  for (byte i = 0; i < uid->size; i++) {
    if (uid->uidByte[i] < 0x10) s += "0";
    s += String(uid->uidByte[i], HEX);
    if (i < uid->size - 1) s += ":";
  }
  s.toUpperCase();
  return s;
}

String readLineWithTimeout(unsigned long timeoutMs) {
  String line = "";
  unsigned long start = millis();
  while (millis() - start < timeoutMs) {
    while (Serial.available() > 0) {
      char c = (char)Serial.read();
      if (c == '\r') continue;
      if (c == '\n') return line;
      line += c;
      if (line.length() > 160) line.remove(0, 80);
    }
  }
  return "";
}

void flushSerialInput() {
  while (Serial.available() > 0) Serial.read();
}

void setup() {
  pinMode(LED_R, OUTPUT);
  pinMode(LED_G, OUTPUT);
  pinMode(LED_B, OUTPUT);

  Serial.begin(115200);
  SPI.begin();
  mfrc522.PCD_Init();

  Wire.begin();
  lcd.init();
  lcd.backlight();

  setColor(0, 0, 1);
  lcdShow("Attendance Sys", "Scan your card");
  Serial.println("ARDUINO_READY");
}

void loop() {
  if (millis() - lastScanMs < SCAN_COOLDOWN_MS) return;

  if (!mfrc522.PICC_IsNewCardPresent()) return;
  if (!mfrc522.PICC_ReadCardSerial()) return;

  lastScanMs = millis();

  String uid = uidToString(&mfrc522.uid);

  flushSerialInput();

  Serial.print("UID,");
  Serial.println(uid);

  lcdShow("Card scanned:", uid);
  delay(600);

  // Countdown
  for (int i = 3; i >= 1; i--) {
    lcdShow("Look to camera", "Starting in " + String(i));
    delay(900);
  }

  lcdShow("Verifying...", "Please hold");
  setColor(0, 0, 1);

  String resp = readLineWithTimeout(PYTHON_TIMEOUT_MS);

  if (resp.startsWith("RESULT,")) {
    String status = resp.substring(7);

    if (status == "CHECKED_VERIFIED") {
      // Success - access granted
      blinkColor(0, 1, 0, 2);
      lcdShow("Access Granted", "Welcome!");
      delay(2000);
    } else if (status == "NOT_AUTHORIZED") {
      // Card not registered or not authorized
      blinkColor(1, 0, 0, 3);
      lcdShow("Not Authorized", "Card not found");
      delay(2000);
    } else if (status == "NO_PHOTO") {
      // No photo in system
      blinkColor(1, 0, 1, 3);
      lcdShow("No Photo", "Contact admin");
      delay(2000);
    } else if (status == "NO_FACE") {
      // Reference photo has no face
      blinkColor(1, 0, 1, 3);
      lcdShow("Photo Error", "Re-register");
      delay(2000);
    } else if (status == "CHECKED_UNVERIFIED") {
      // Face didn't match
      blinkColor(1, 0, 0, 3);
      lcdShow("Face Mismatch", "Try again");
      delay(2000);
    } else if (status == "ERROR") {
      // System error
      blinkColor(1, 1, 0, 3);
      lcdShow("System Error", "Try again");
      delay(2000);
    } else {
      // Unknown response
      blinkColor(1, 0, 0, 2);
      lcdShow("Unknown Error", status);
      delay(2000);
    }
  } else {
    // No response - timeout
    blinkColor(1, 0, 0, 3);
    lcdShow("Timeout", "No response");
    delay(2000);
  }

  setColor(0, 0, 1);
  lcdShow("Attendance Sys", "Scan your card");

  mfrc522.PICC_HaltA();
  mfrc522.PCD_StopCrypto1();
}
