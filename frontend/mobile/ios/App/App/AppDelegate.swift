import Capacitor
import UIKit
import WebKit

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {

    var window: UIWindow?

    func application(
        _ app: UIApplication, open url: URL, options: [UIApplication.OpenURLOptionsKey: Any] = [:]
    ) -> Bool {
        print("✅ AppDelegate: Received URL - \(url.absoluteString)")

        if url.scheme == "savetoink", url.host == "share",
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
            let articleUrl = components.queryItems?.first(where: { $0.name == "url" })?.value,
            let encoded = articleUrl.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed),
            let rootVC = self.window?.rootViewController as? RootViewController,
            !RootViewController.appUrl.isEmpty,
            let destination = URL(string: "\(RootViewController.appUrl)/new?url=\(encoded)")
        {
            DispatchQueue.main.async {
                rootVC.bridge?.webView?.load(URLRequest(url: destination))
            }
            return true
        }

        return ApplicationDelegateProxy.shared.application(app, open: url, options: options)
    }
}

class RootViewController: CAPBridgeViewController, WKNavigationDelegate, WKUIDelegate {

    private var errorView: UIView?
    private var errorLabel: UILabel?
    private var retryButton: UIButton?

    private static var allowedDomains: [String] = []
    private static let capacitorLocalhost = "capacitor://localhost"
    static var appUrl: String = ""

    override func viewDidLoad() {
        super.viewDidLoad()

        loadAllowedDomains()

        if let webView = self.bridge?.webView {
            webView.navigationDelegate = self
            webView.uiDelegate = self
        }
    }

    private func loadAllowedDomains() {
        guard Self.allowedDomains.isEmpty else { return }

        guard let configPath = Bundle.main.path(forResource: "capacitor.config", ofType: "json"),
            let data = try? Data(contentsOf: URL(fileURLWithPath: configPath)),
            let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
            let server = json["server"] as? [String: Any],
            let allowNavigation = server["allowNavigation"] as? [String]
        else {
            print("❌ Failed to load capacitor.config.json")
            return
        }

        Self.allowedDomains = allowNavigation.compactMap { url in
            guard let parsed = URL(string: url) else { return nil }
            return parsed.host
        }

        Self.appUrl = (server["url"] as? String) ?? ""

        print("✅ Loaded allowed domains: \(Self.allowedDomains), appUrl: \(Self.appUrl)")
    }

    override func viewSafeAreaInsetsDidChange() {
        super.viewSafeAreaInsetsDidChange()
        adjustWebViewForSafeArea()
    }

    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        adjustWebViewForSafeArea()
    }

    private func adjustWebViewForSafeArea() {
        guard let webView = self.bridge?.webView else { return }

        let safeAreaTop = view.safeAreaInsets.top
        let statusBarHeight = view.window?.windowScene?.statusBarManager?.statusBarFrame.height ?? 0

        if safeAreaTop > 0 || statusBarHeight > 0 {
            let topInset = max(safeAreaTop, statusBarHeight)
            print("📱 Adjusting webview for safe area - Top inset: \(topInset)")

            let scrollView = webView.scrollView
            scrollView.contentInset = UIEdgeInsets(top: topInset, left: 0, bottom: 0, right: 0)
            scrollView.scrollIndicatorInsets = scrollView.contentInset
        }
    }

    func webView(
        _ webView: WKWebView,
        didReceiveServerRedirectForProvisionalNavigation _: WKNavigation!
    ) {
        if let url = webView.url {
            print("🔄 Server redirect to: \(url.absoluteString)")
        }
    }

    func webView(
        _ webView: WKWebView, createWebViewWith _: WKWebViewConfiguration,
        for navigationAction: WKNavigationAction, windowFeatures _: WKWindowFeatures
    ) -> WKWebView? {
        guard let url = navigationAction.request.url else { return nil }

        let urlString = url.absoluteString
        print("🪟 New window navigation to: \(urlString)")

        let host = url.host ?? ""
        let isHTTP = url.scheme == "http" || url.scheme == "https"

        if isHostAllowed(host) || isLocalhost(host) {
            print("✅ Loading new window in current webview: \(urlString)")
            webView.load(URLRequest(url: url))
        } else if isHTTP {
            print("🔗 Opening external URL in Safari: \(urlString)")
            UIApplication.shared.open(url, options: [:], completionHandler: nil)
        } else {
            print("✅ Allowing new window: \(urlString)")
        }

        return nil
    }

    func webViewDidClose(_: WKWebView) {
        print("🪟 Webview closed")
    }

    func webView(
        _: WKWebView, decidePolicyFor navigationAction: WKNavigationAction,
        decisionHandler: @escaping (WKNavigationActionPolicy) -> Void
    ) {
        guard let url = navigationAction.request.url else {
            decisionHandler(.allow)
            return
        }

        let urlString = url.absoluteString
        print(
            "🔍 Navigation action type: \(navigationAction.navigationType.rawValue), URL: \(urlString)"
        )

        let host = url.host ?? ""
        let isFileURL = url.scheme == "file"
        let isAboutBlank = urlString.contains("about:blank")
        let isCapacitorJS = urlString.contains("capacitor.js")
        let isCapacitorLocalhost = urlString.contains(Self.capacitorLocalhost)
        let isHTTP = url.scheme == "http" || url.scheme == "https"

        if isHostAllowed(host) || isFileURL || isAboutBlank || isCapacitorJS
            || isCapacitorLocalhost || isLocalhost(host)
        {
            print("✅ Allowing navigation to: \(urlString)")
            decisionHandler(.allow)
        } else if isHTTP {
            print("🔗 Opening external URL in Safari: \(urlString)")
            UIApplication.shared.open(url, options: [:], completionHandler: nil)
            decisionHandler(.cancel)
        } else {
            // mailto:, tel:, etc — let the system handle them
            print("🔗 Opening system URL: \(urlString)")
            UIApplication.shared.open(url, options: [:], completionHandler: nil)
            decisionHandler(.cancel)
        }
    }

    func webView(
        _: WKWebView, decidePolicyFor navigationResponse: WKNavigationResponse,
        decisionHandler: @escaping (WKNavigationResponsePolicy) -> Void
    ) {
        if let url = navigationResponse.response.url {
            print("📥 Navigation response for: \(url.absoluteString)")
        }
        decisionHandler(.allow)
    }

    func webView(
        _: WKWebView,
        didFailProvisionalNavigation _: WKNavigation!,
        withError error: Error
    ) {
        let nsError = error as NSError
        print("❌ Provisional navigation failed with error: \(nsError.localizedDescription)")
        showErrorIfNetworkError(nsError)
    }

    func webView(_: WKWebView, didFail _: WKNavigation!, withError error: Error) {
        let nsError = error as NSError
        print("❌ Navigation failed with error: \(nsError.localizedDescription)")
        showErrorIfNetworkError(nsError)
    }

    func webView(_: WKWebView, didFinish _: WKNavigation!) {
        hideError()
    }

    private func isHostAllowed(_ host: String) -> Bool {
        Self.allowedDomains.contains(host) || Self.allowedDomains.contains { host.hasSuffix($0) }
    }

    private func isLocalhost(_ host: String) -> Bool {
        host == "localhost" || host == "127.0.0.1" || host == "::1"
    }

    private func showErrorIfNetworkError(_ error: NSError) {
        let networkErrorCodes: Set<Int> = [
            NSURLErrorNotConnectedToInternet,
            NSURLErrorTimedOut,
            NSURLErrorCannotConnectToHost,
            NSURLErrorNetworkConnectionLost,
            NSURLErrorDNSLookupFailed,
            NSURLErrorResourceUnavailable,
            NSURLErrorCannotFindHost,
        ]

        if networkErrorCodes.contains(error.code) {
            showError()
        }
    }

    private func showError() {
        guard errorView == nil, self.bridge?.webView != nil else { return }

        let errorView = UIView()
        errorView.backgroundColor = UIColor.systemBackground
        errorView.translatesAutoresizingMaskIntoConstraints = false
        self.view.addSubview(errorView)
        self.errorView = errorView

        NSLayoutConstraint.activate([
            errorView.topAnchor.constraint(equalTo: self.view.topAnchor),
            errorView.bottomAnchor.constraint(equalTo: self.view.bottomAnchor),
            errorView.leadingAnchor.constraint(equalTo: self.view.leadingAnchor),
            errorView.trailingAnchor.constraint(equalTo: self.view.trailingAnchor),
        ])

        let stackView = UIStackView()
        stackView.axis = .vertical
        stackView.spacing = 20
        stackView.alignment = .center
        stackView.translatesAutoresizingMaskIntoConstraints = false
        errorView.addSubview(stackView)

        NSLayoutConstraint.activate([
            stackView.centerXAnchor.constraint(equalTo: errorView.centerXAnchor),
            stackView.centerYAnchor.constraint(equalTo: errorView.centerYAnchor),
            stackView.leadingAnchor.constraint(
                greaterThanOrEqualTo: errorView.leadingAnchor, constant: 20),
            stackView.trailingAnchor.constraint(
                lessThanOrEqualTo: errorView.trailingAnchor, constant: -20),
        ])

        let errorLabel = UILabel()
        errorLabel.text = "Save to Ink backend at \(Self.appUrl) temporarily unavailable"
        errorLabel.textAlignment = .center
        errorLabel.numberOfLines = 0
        errorLabel.font = UIFont.systemFont(ofSize: 18, weight: .medium)
        errorLabel.translatesAutoresizingMaskIntoConstraints = false
        stackView.addArrangedSubview(errorLabel)
        self.errorLabel = errorLabel

        let retryButton = UIButton(type: .system)
        retryButton.setTitle("Retry", for: .normal)
        retryButton.titleLabel?.font = UIFont.systemFont(ofSize: 16, weight: .medium)
        retryButton.translatesAutoresizingMaskIntoConstraints = false
        retryButton.addTarget(self, action: #selector(retryTapped), for: .touchUpInside)
        stackView.addArrangedSubview(retryButton)
        self.retryButton = retryButton

        NSLayoutConstraint.activate([
            retryButton.widthAnchor.constraint(greaterThanOrEqualToConstant: 120),
            retryButton.heightAnchor.constraint(equalToConstant: 44),
        ])
    }

    private func hideError() {
        errorView?.removeFromSuperview()
        errorView = nil
        errorLabel = nil
        retryButton = nil
    }

    @objc private func retryTapped() {
        // Navigate to app URL directly instead of reloading
        // Error view stays visible until navigation succeeds (didFinish hides it)
        if let appUrl = URL(string: Self.appUrl) {
            self.bridge?.webView?.load(URLRequest(url: appUrl))
        }
    }
}
