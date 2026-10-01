package dev.narehatechi.narecord.mobile

import android.app.Activity
import android.os.Bundle
import android.view.Gravity
import android.widget.TextView

class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        setContentView(
            TextView(this).apply {
                text = getString(R.string.preview_message)
                textSize = 18f
                gravity = Gravity.CENTER
                setPadding(32, 32, 32, 32)
            },
        )
    }
}
