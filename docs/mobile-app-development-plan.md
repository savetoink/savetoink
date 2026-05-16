# Mobile Application Development Plan

## Overview

A minimal native app shell built with [CapacitorJS](https://capacitorjs.com/) that does exactly two things:

1. **Registers as a share target** — appears in the system share sheet
2. **Redirects shared URLs** — opens `<APP_URL>/new?url=<shared_url>` in the system browser

No WebView, no embedded UI, no session management. The app is a thin bridge between the OS share sheet and the existing webapp.

All share handling is done in **pure native Swift** code — the Capacitor JS layer is intentionally empty. The native `SceneDelegate` handles the custom URL scheme and opens Safari directly.

## Goals

- **Minimal App**: No WebView, no embedded UI — just a share receiver
- **Zero Auth Complexity**: Auth stays in the browser where it already works
- **Simple Maintenance**: No webapp code is bundled; only the share bridge is native
- **Incoming Share Only**: Users share links from any app → SaveToInk opens in browser

## Technology Stack

- **Package Manager**: Bun
- **Framework**: CapacitorJS 8.x (for native project management only; no JS plugins used)
- **Incoming Share**: Native iOS Share Extension (`ShareViewController.swift`)
- **Outgoing**: Native `UIApplication.shared.open()` to launch the system browser
- **URL scheme**: Custom `savetoink://` for ShareExtension → AppDelegate communication
- **App Group**: `group.ink.saveto.app` for ShareExtension ↔ AppDelegate data sharing
- **Platforms**: iOS first, Android later

## How It Works

```
User shares URL from any app
  → OS share sheet shows "Save to Ink"
  → ShareExtension saves URL to App Group UserDefaults
  → ShareExtension opens main app via savetoink://share
  → SceneDelegate intercepts savetoink:// URL scheme
  → Reads shared URL from App Group UserDefaults
  → Opens system browser at <APP_URL>/new?url=<shared_url>
```

---

## 1. Project Setup — ✅ COMPLETED

### 1.1 File Structure

```
frontend/mobile/
├── capacitor.config.ts        # App config (appId: ink.saveto.app), SPM setup
├── index.html                 # Entry HTML (thin wrapper, src/index.ts is empty)
├── package.json               # Capacitor 8.x deps, no share/browser plugins
├── vite.config.ts             # Minimal build config, no env injection
├── icon-targets.csv            # Icon generation targets
├── tsconfig.json
├── public/                    # App icons
├── src/
│   ├── index.ts               # Intentionally empty — share handling is native
│   └── vite-env.d.ts
├── ios/
│   ├── debug.xcconfig          # CAPACITOR_DEBUG = true
│   ├── .gitignore
│   └── App/
│       ├── App.xcodeproj/
│       ├── App/
│       │   ├── AppDelegate.swift       # App lifecycle, scene configuration
│       │   ├── SceneDelegate.swift     # Handles savetoink://share, opens Safari
│       │   ├── App.entitlements        # App Group: group.ink.saveto.app
│       │   ├── Info.plist              # URL scheme: savetoink, APP_URL build setting
│       │   ├── Assets.xcassets/
│       │   ├── Base.lproj/
│       │   ├── capacitor.config.json
│       │   └── config.xml
│       ├── ShareExtension/
│       │   ├── ShareViewController.swift  # Saves to UserDefaults, opens main app
│       │   ├── ShareExtension.entitlements  # App Group: group.ink.saveto.app
│       │   ├── Info.plist                 # NSExtension activation rules
│       │   └── Base.lproj/MainInterface.storyboard
│       └── CapApp-SPM/                   # Swift Package Manager config for Capacitor
```

### 1.2 Key Source Files

**`src/index.ts`** — Intentionally empty; share handling is entirely native:
```typescript
// Capacitor requires webDir to contain an index.html with a JS entry point.
// Since share handling is done entirely in native code (SceneDelegate.swift),
// this file is intentionally empty.
```

**`capacitor.config.ts`** — No plugin config (share handling is native):
```typescript
import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "ink.saveto.app",
  appName: "Save to Ink",
  webDir: "dist",
  loggingBehavior: "debug",
  experimental: {
    ios: {
      spm: {
        swiftToolsVersion: "6.2",
      },
    },
  },
};

export default config;
```

**`vite.config.ts`** — Minimal build config; no `PUBLIC_APP_URL` injection (URL comes from Xcode build settings):
```typescript
import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  build: {
    target: "esnext",
    outDir: "dist",
    emptyOutDir: true,
  },
});
```

**`AppDelegate.swift`** — App lifecycle and scene configuration (UIScene lifecycle):
```swift
import UIKit
import Capacitor

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {

    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        return true
    }

    func application(_ application: UIApplication, configurationForConnecting connectingSceneSession: UISceneSession, options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        return UISceneConfiguration(name: "Default Configuration", sessionRole: connectingSceneSession.role)
    }

    func application(_ application: UIApplication, didDiscardSceneSessions sceneSessions: Set<UISceneSession>) {
    }
}
```

**`SceneDelegate.swift`** — URL scheme handling and cold-start redirect:
```swift
import UIKit

class SceneDelegate: UIResponder, UISceneDelegate {

    var window: UIWindow?
    private var didHandleInitialRedirect = false

    func scene(_ scene: UIScene, willConnectTo session: UISceneSession, options connectionOptions: UIScene.ConnectionOptions) {
        // Handle URL if app was launched via URL scheme (e.g. savetoink://share)
        if let urlContext = connectionOptions.urlContexts.first {
            didHandleInitialRedirect = true
            handleURL(urlContext.url)
        }
    }

    func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
        // Handle URL when app is already running
        didHandleInitialRedirect = true
        if let urlContext = URLContexts.first {
            handleURL(urlContext.url)
        }
    }

    func sceneDidBecomeActive(_ scene: UIScene) {
        // On cold start (tap app icon), redirect to APP_URL immediately
        guard !didHandleInitialRedirect else { return }
        didHandleInitialRedirect = true
        openSafari(path: "")
    }

    // MARK: - Share handling
    // handleSharedContent() and openSafari(path:) — same logic as before
}
```

**`Info.plist` (App)** — Scene manifest, custom URL scheme, and `APP_URL` build setting:
```xml
<key>UIApplicationSceneManifest</key>
<dict>
    <key>UIApplicationSupportsMultipleScenes</key>
    <false/>
    <key>UISceneConfigurations</key>
    <dict>
        <key>UIWindowSceneSessionRoleApplication</key>
        <array>
            <dict>
                <key>UISceneConfigurationName</key>
                <string>Default Configuration</string>
                <key>UISceneDelegateClassName</key>
                <string>$(PRODUCT_MODULE_NAME).SceneDelegate</string>
                <key>UISceneStoryboardFile</key>
                <string>Main</string>
            </dict>
        </array>
    </dict>
</dict>
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>savetoink</string>
        </array>
    </dict>
</array>
<key>APP_URL</key>
<string>$(APP_URL)</string>
```

### 1.3 Dependencies (`package.json`)

```json
{
  "dependencies": {
    "@capacitor/android": "^8.2.0",
    "@capacitor/cli": "^8.2.0",
    "@capacitor/core": "^8.2.0",
    "@capacitor/ios": "^8.2.0"
  },
  "devDependencies": {
    "@types/bun": "latest",
    "vite": "^8.0.0"
  }
}
```

Note: No `@capgo/capacitor-share-target` or `@capacitor/browser` — all share handling is native Swift.

---

## 2. iOS Configuration — ✅ COMPLETED

### 2.1 App Groups
- ✅ `App/App.entitlements` — `group.ink.saveto.app`
- ✅ `ShareExtension/ShareExtension.entitlements` — `group.ink.saveto.app`

### 2.2 URL Scheme
- ✅ `App/Info.plist` — `CFBundleURLSchemes` → `savetoink`
- Share extension opens main app via `savetoink://share`
- SceneDelegate intercepts `savetoink://` scheme, reads UserDefaults, opens Safari

### 2.3 APP_URL Build Setting
- ✅ `App/Info.plist` — `APP_URL` key set to `$(APP_URL)` (Xcode build setting)
- Must be configured per-environment (e.g., dev/staging/production URLs)

#### How APP_URL is injected

The app uses a **three-step pipeline** with no Vite involvement at runtime:

```
project.pbxproj (build setting)     debug.xcconfig
  APP_URL = "https://..."            CAPACITOR_DEBUG = true
         │                                    │
         └──────────┬─────────────────────────┘
                    ▼  $(VAR) substitution at build time
              Info.plist (in .app bundle)
                APP_URL → "https://..."
                CAPACITOR_DEBUG → "true"
                    │
                    ▼  Bundle.main.object(forInfoDictionaryKey:)
              SceneDelegate.swift (runtime)
```

1. **Xcode build setting** (`project.pbxproj`) — `APP_URL` is defined as a user-defined build setting
2. **Info.plist** — references it via `$(APP_URL)` substitution; Xcode resolves it at build time
3. **SceneDelegate.swift** — reads it at runtime via `Bundle.main.object(forInfoDictionaryKey: "APP_URL")`

#### Recommended: `.xcconfig` files per environment

Currently `APP_URL` is hardcoded in `project.pbxproj` for both Debug and Release. The recommended approach is to move it into `.xcconfig` files so CI/CD can override it per-environment without modifying `pbxproj`:

**`debug.xcconfig`** (already exists for `CAPACITOR_DEBUG`):
```
CAPACITOR_DEBUG = true
APP_URL = https://app.dev.saveto.ink
```

**`release.xcconfig`** (create for production):
```
APP_URL = https://app.saveto.ink
```

Then assign them in Xcode: **Project → Info → Configurations** — set each build configuration to use its respective `.xcconfig` file as the `baseConfigurationReference`.

For CI/CD, a pre-build script can replace placeholder values in the `.xcconfig` files with actual environment variables:
```bash
#!/bin/bash
# Example: inject APP_URL from CI env var into the xcconfig
sed -i '' "s/APP_URL = .*/APP_URL = ${APP_URL}/" "$PROJECT_DIR/frontend/mobile/ios/debug.xcconfig"
```

**References:**
- [Adding a build configuration file to your project](https://developer.apple.com/documentation/xcode/adding-a-build-configuration-file-to-your-project) — Apple's official `.xcconfig` docs
- [Customizing the build schemes for a project](https://developer.apple.com/documentation/xcode/customizing-the-build-schemes-for-a-project) — Scheme-level env vars for debugging
- [Let's Set Up Your iOS Environments](https://thoughtbot.com/blog/let-s-setup-your-ios-environments) — comprehensive `.xcconfig` per-environment walkthrough
- [Securely Manage iOS App Data with .xcconfig and CI/CD](https://povio.com/blog/securely-manage-sensitive-data-with-xcconfig-files-and-ci-cd-injection) — CI/CD injection pattern

### 2.4 Share Extension
- ✅ `ShareExtension/ShareViewController.swift` — extracts URLs/text, saves to App Group UserDefaults, opens main app via custom URL scheme
- ✅ `ShareExtension/Info.plist` — `NSExtensionActivationRule` restricted to URLs and text only (`NSExtensionActivationSupportsWebURLWithMaxCount: 1`, `NSExtensionActivationSupportsText: true`)
- ✅ Deployment target matched to iOS 15.0

### 2.5 Xcode Project
- ✅ ShareExtension target added as dependency of App target
- ✅ ShareExtension embedded via "Embed Foundation Extensions" phase
- ✅ Code signing configured (team: `WZA32485UK`)
- ✅ Bundle IDs: `ink.saveto.app` (app), `ink.saveto.app.ShareExtension` (extension)

---

## 3. Testing — 🔲 TODO

### Build & Run
```bash
cd frontend/mobile
bun run build
bunx cap sync ios
bunx cap open ios
# In Xcode: APP_URL is read from the .xcconfig file assigned to the active build configuration
# To override locally, edit debug.xcconfig or set APP_URL in the scheme's environment variables
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
- [ ] Set `APP_URL` build setting for production
- [ ] Build release archive
- [ ] Create App Store Connect listing (description, keywords, screenshots)
- [ ] Submit to TestFlight for internal testing
- [ ] Submit for App Store review

---

## 5. Android — 🔲 DEFERRED

- [ ] Run `bunx cap add android`
- [ ] Configure Android share intent (native Java/Kotlin, no JS bridge)
- [ ] Test on Android devices
- [ ] Submit to Google Play

---

## 6. References

- [Capacitor Documentation](https://capacitorjs.com/docs)
- [iOS Share Extensions](https://developer.apple.com/documentation/uikit/uiactivityviewcontroller)
- [UIApplication.open](https://developer.apple.com/documentation/uikit/uiapplication/1648685-open)
- [UserDefaults Suite](https://developer.apple.com/documentation/foundation/userdefaults/1409957-init)
