import UIKit
import UniformTypeIdentifiers

class ShareViewController: UIViewController {

    override func viewDidAppear(_ animated: Bool) {
        super.viewDidAppear(animated)

        guard let item = extensionContext?.inputItems.first as? NSExtensionItem,
            let provider = item.attachments?.first
        else {
            done()
            return
        }

        // Handle a shared URL
        if provider.hasItemConformingToTypeIdentifier(UTType.url.identifier) {
            provider.loadItem(forTypeIdentifier: UTType.url.identifier) { [weak self] data, _ in
                let url: URL? =
                    data as? URL
                    ?? (data as? String).flatMap(URL.init)
                self?.open(sharedUrl: url?.absoluteString ?? "")
            }
            return
        }

        // Handle plain text (some apps share URLs as text)
        if provider.hasItemConformingToTypeIdentifier(UTType.plainText.identifier) {
            provider.loadItem(forTypeIdentifier: UTType.plainText.identifier) {
                [weak self] data, _ in
                let text = data as? String ?? ""
                self?.open(sharedUrl: text)
            }
            return
        }

        done()
    }

    private func open(sharedUrl: String) {
        guard !sharedUrl.isEmpty,
            let encoded = sharedUrl.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed),
            let appUrl = URL(string: "savetoink://share?url=\(encoded)")
        else {
            done()
            return
        }

        // UIApplication.shared is unavailable in extensions, so walk the responder chain
        var responder: UIResponder? = self
        while responder != nil {
            if let app = responder as? UIApplication {
                app.open(appUrl)
                break
            }
            responder = responder?.next
        }

        done()
    }

    private func done() {
        extensionContext?.completeRequest(returningItems: [], completionHandler: nil)
    }
}
