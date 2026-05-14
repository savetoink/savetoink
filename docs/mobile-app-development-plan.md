# Mobile Application Development Plan

## Overview

A minimal native app shell built with [CapacitorJS](https://capacitorjs.com/) that does exactly two things:

1. **Registers as a share target** — appears in the system share sheet
2. **Redirects shared URLs** — opens `<PUBLIC_APP_URL>/new?url=<shared_url>` in the system browser

No WebView, no embedded UI, no session management. The app is a thin bridge between the OS share sheet and the existing webapp.

## Goals

- **Minimal App**: No WebView, no embedded UI — just a share receiver
- **Zero Auth Complexity**: Auth stays in the browser where it already works
- **Simple Maintenance**: No webapp code is bundled; only the share bridge is native
- **Incoming Share Only**: Users share links from any app → SaveToInk opens in browser

## Technology Stack

- **Package Manager**: Bun
- **Framework**: CapacitorJS 8.x
- **Incoming Share**: [`@capgo/capacitor-share-target`](https://github.com/Cap-go/capacitor-share-target) plugin
- **Outgoing**: [`@capacitor/browser`](https://capacitorjs.com/docs/apis/browser) to open the system browser
- **Platforms**: iOS first, Android later

## How It Works

```
User shares URL from any app
  → OS share sheet shows "Save to Ink"
  → ShareExtension saves URL to App Group UserDefaults
  → ShareExtension opens main app via savetoink://share
  → CapacitorShareTarget plugin reads UserDefaults
  → Fires "shareReceived" JS event
  → src/index.ts calls Browser.open(<APP_URL>/new?url=<shared_url>)
```

---

## 1. Project Setup — ✅ COMPLETED

### 1.1 File Structure

```
frontend/mobile/
├── capacitor.config.ts       # App config (appId: ink.saveto.app)
├── index.html                # Entry HTML
├── package.json              # Capacitor 8.x deps
├── vite.config.ts            # PUBLIC_APP_URL injection
├── public/                   # App icons
├── src/
│   ├── index.ts              # Share listener → Browser.open()
│   └── vite-env.d.ts
└── ios/
    └── App/
        ├── App/
        │   ├── AppDelegate.swift
        │   ├── App.entitlements          # App Group: group.ink.saveto.app
        │   └── Info.plist                # URL scheme: savetoink
        ├── ShareExtension/
        │   ├── ShareViewController.swift # Saves to UserDefaults, opens main app
        │   ├── ShareExtension.entitlements  # App Group: group.ink.saveto.app
        │   └── Info.plist                # NSExtension activation rules
        └── App.xcodeproj/
```

### 1.2 Key Source Files

**`src/index.ts`** — The entire app logic:
```typescript
import { CapacitorShareTarget } from "@capgo/capacitor-share-target";
import { Browser } from "@capacitor/browser";
import { Capacitor } from "@capacitor/core";

const APP_URL = import.meta.env.PUBLIC_APP_URL;

if (Capacitor.isNativePlatform()) {
  CapacitorShareTarget.addListener("shareReceived", async (event) => {
    const url = event.texts?.[0];
    if (!url) return;
    if (!url.startsWith("http://") && !url.startsWith("https://")) return;
    await Browser.open({ url: `${APP_URL}/new?url=${encodeURIComponent(url)}` });
  });
}
```

**`capacitor.config.ts`**:
```typescript
import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "ink.saveto.app",
  appName: "Save to Ink",
  webDir: "dist",
  loggingBehavior: "debug",
  plugins: {
    CapacitorShareTarget: {
      appGroupId: "group.ink.saveto.app",
    },
  },
};

export default config;
```

---

## 2. iOS Configuration — ✅ COMPLETED

### 2.1 App Groups
- ✅ `App/App.entitlements` — `group.ink.saveto.app`
- ✅ `ShareExtension/ShareExtension.entitlements` — `group.ink.saveto.app`

### 2.2 URL Scheme
- ✅ `App/Info.plist` — `CFBundleURLSchemes` → `savetoink`
- Share extension opens main app via `savetoink://share`

### 2.3 Share Extension
- ✅ `ShareExtension/ShareViewController.swift` — extracts URLs/text, saves to App Group UserDefaults, opens main app
- ✅ `ShareExtension/Info.plist` — `NSExtensionActivationRule` restricted to URLs and text only (`NSExtensionActivationSupportsWebURLWithMaxCount: 1`, `NSExtensionActivationSupportsText: true`)
- ✅ Deployment target matched to iOS 15.0

### 2.4 Xcode Project
- ✅ ShareExtension target added as dependency of App target
- ✅ ShareExtension embedded via "Embed Foundation Extensions" phase
- ✅ Code signing configured (team: `WZA32485UK`)
- ✅ Bundle IDs: `ink.saveto.app` (app), `ink.saveto.app.ShareExtension` (extension)

---

## 3. Testing — 🔲 TODO

### Build & Run
```bash
cd frontend/mobile
PUBLIC_APP_URL=https://your-app-url bun run build
bunx cap sync ios
bunx cap open ios
# Run on simulator/device from Xcode
```

### Test Checklist
- [ ] Build succeeds in Xcode
- [ ] App launches on simulator
- [ ] Share a URL from Safari → "Save to Ink" appears in share sheet
- [ ] Selecting it opens the browser at `/new?url=...`
- [ ] Share a URL from other apps (Notes, Messages)
- [ ] Share non-URL text → ignored gracefully
- [ ] Share when app is in background
- [ ] Share when app is not running (cold start)
- [ ] Auth0 login works in the opened browser
- [ ] Full save flow works (share → login → save)

---

## 4. App Store Submission — 🔲 TODO

- [ ] Configure production code signing
- [ ] Build release archive
- [ ] Create App Store Connect listing (description, keywords, screenshots)
- [ ] Submit to TestFlight for internal testing
- [ ] Submit for App Store review

---

## 5. Android — 🔲 DEFERRED

- [ ] Run `bunx cap add android`
- [ ] Configure Android share intent
- [ ] Test on Android devices
- [ ] Submit to Google Play

---

## 6. References

- [Capacitor Documentation](https://capacitorjs.com/docs)
- [@capgo/capacitor-share-target](https://github.com/Cap-go/capacitor-share-target)
- [@capacitor/browser](https://capacitorjs.com/docs/apis/browser)
- [iOS Share Extensions](https://developer.apple.com/documentation/uikit/uiactivityviewcontroller)
