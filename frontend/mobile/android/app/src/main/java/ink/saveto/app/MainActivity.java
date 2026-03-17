package ink.saveto.app;

import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        handleShareIntent(getIntent());
    }

    @Override
    public void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        handleShareIntent(intent);
    }

    private void handleShareIntent(Intent intent) {
        if (intent == null) {
            return;
        }

        String action = intent.getAction();
        String type = intent.getType();

        if (Intent.ACTION_SEND.equals(action) && type != null) {
            String sharedUrl = null;

            if ("text/uri-list".equals(type) || "text/plain".equals(type)) {
                Uri uri = intent.getParcelableExtra(Intent.EXTRA_STREAM);
                if (uri != null) {
                    sharedUrl = uri.toString();
                } else {
                    sharedUrl = intent.getStringExtra(Intent.EXTRA_TEXT);
                }
            }

            if (sharedUrl != null) {
                String encodedUrl = Uri.encode(sharedUrl);
                String targetUrl = BuildConfig.APP_URL + "/new?url=" + encodedUrl;

                if (getBridge() != null && getBridge().getWebView() != null) {
                    getBridge().getWebView().loadUrl(targetUrl);
                }
            }
        }
    }
}
