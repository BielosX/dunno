locals {
  query = <<-EOF
  SOURCE logGroups(namePrefix: ["/aws/lambda/dunno"], class: "STANDARD")
  | fields @timestamp, @message, level
  | filter level = "error"
  | sort @timestamp desc
  | limit 10000
  EOF
}

resource "aws_cloudwatch_dashboard" "dashboard" {
  dashboard_name = local.prefix
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 6
        height = 6
        properties = {
          metrics = [
            ["dunno", "requestDuration", "pattern", "/books/", "codeFamily", "2XX",
              { region : local.region }
            ]
          ]
          view    = "timeSeries"
          stacked = false
          region  = local.region
          title   = "API requestDuration"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type   = "metric"
        x      = 6
        y      = 0,
        width  = 6
        height = 6
        properties = {
          metrics = [
            ["dunno", "openSearchQueryDuration", "params", "title",
              { region : local.region }
            ]
          ]
          view    = "timeSeries"
          stacked = false
          region  = local.region
          title   = "openSearchQueryDuration"
          period  = 60
          stat    = "Average"
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 24
        height = 6
        properties = {
          query         = local.query
          queryLanguage = "CWLI"
          queryBy       = "logGroupPrefix"
          logGroupPrefixes = {
            logClass = "STANDARD"
            logGroupPrefix = [
              "/aws/lambda/dunno"
            ],
            accountIds = [
              "All"
            ]
          },
          region = local.region,
          title  = "Dunno Errors",
          view   = "table"
        }
      }
    ]
  })
}