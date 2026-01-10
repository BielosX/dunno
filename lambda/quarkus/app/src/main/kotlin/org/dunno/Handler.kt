package org.dunno

import com.amazonaws.services.lambda.runtime.Context
import com.amazonaws.services.lambda.runtime.RequestHandler

class Handler : RequestHandler<Void, String> {
  override fun handleRequest(unused: Void?, context: Context?): String {
    return context?.logStreamName ?: "unknown"
  }
}
