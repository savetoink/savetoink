import UIKit
import Capacitor

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {

    var window: UIWindow?
    private var didHandleInitialRedirect = false

    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        return true
    }

    func application(_ app: UIApplication, open url: URL, options: [UIApplication.OpenURLOptionsKey: Any] = [:]) -> Bool {
        // App was opened via URL scheme — intent is already handled, skip initial redirect
        didHandleInitialRedirect = true
        if url.scheme == "savetoink" {
            handleSharedContent()
            return true
        }
        return ApplicationDelegateProxy.shared.application(app, open: url, options: options)
    }

    func applicationDidBecomeActive(_ application: UIApplication) {
        // On cold start (tap app icon), redirect to APP_URL immediately
        guard !didHandleInitialRedirect else { return }
        didHandleInitialRedirect = true
        openSafari(path: "")
    }

    func application(_ application: UIApplication, continue userActivity: NSUserActivity, restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void) -> Bool {
        return ApplicationDelegateProxy.shared.application(application, continue: userActivity, restorationHandler: restorationHandler)
    }

    // MARK: - Share handling

    private func handleSharedContent() {
        let appGroup = "group.ink.saveto.app"
        let userDefaults = UserDefaults(suiteName: appGroup)

        guard let shareData = userDefaults?.dictionary(forKey: "share-target-data"),
              let texts = shareData["texts"] as? [String],
              let sharedUrl = texts.first,
              sharedUrl.hasPrefix("http") else {
            print("[App] No valid shared URL found, opening webapp home")
            openSafari(path: "")
            return
        }

        // Clear consumed data
        userDefaults?.removeObject(forKey: "share-target-data")

        let encoded = sharedUrl.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? sharedUrl
        openSafari(path: "/new?url=\(encoded)")
    }

    private func openSafari(path: String) {
        guard let appUrl = Bundle.main.object(forInfoDictionaryKey: "APP_URL") as? String,
              !appUrl.isEmpty else {
            print("[App] ERROR: APP_URL not configured in Info.plist")
            return
        }
        let fullUrl = "\(appUrl)\(path)"
        guard let url = URL(string: fullUrl) else {
            print("[App] ERROR: Invalid URL: \(fullUrl)")
            return
        }
        UIApplication.shared.open(url, options: [:], completionHandler: nil)
    }
}
