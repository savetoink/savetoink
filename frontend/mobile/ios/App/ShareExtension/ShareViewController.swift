import UIKit
import UniformTypeIdentifiers

class ShareViewController: UIViewController {

    let APP_GROUP_ID = "group.ink.saveto.app"

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
                self?.saveAndOpen(text: url?.absoluteString ?? "")
            }
            return
        }

        // Handle plain text (some apps share URLs as text)
        if provider.hasItemConformingToTypeIdentifier(UTType.plainText.identifier) {
            provider.loadItem(forTypeIdentifier: UTType.plainText.identifier) {
                [weak self] data, _ in
                let text = data as? String ?? ""
                self?.saveAndOpen(text: text)
            }
            return
        }

        done()
    }

    private func saveAndOpen(text: String) {
        guard !text.isEmpty else {
            done()
            return
        }

        // Save to shared UserDefaults so the CapacitorShareTarget plugin can read it
        let shareData: [String: Any] = [
            "title": "",
            "texts": [text],
            "files": [],
        ]

        let userDefaults = UserDefaults(suiteName: APP_GROUP_ID)
        userDefaults?.set(shareData, forKey: "share-target-data")
        userDefaults?.synchronize()

        // Open the main app via URL scheme so it becomes active and the plugin can process
        guard let appUrl = URL(string: "savetoink://share") else {
            done()
            return
        }

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
