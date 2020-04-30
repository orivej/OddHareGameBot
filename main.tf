variable "name" {
  type        = string
  default     = "enlapin" # basename(path.root)
  description = "Name of the binary (basename of this directory)"
}
variable "debug" {
  type        = bool
  default     = false
  description = "Debug logging"
}
variable "token" {
  type        = string
  description = "Telegram bot token"
}
variable "region" {
  type        = string
  description = "AWS deployment region"
}

locals {
  exe      = var.name
  zip      = "${local.exe}.zip"
  table    = "${var.name}Table"
  func     = "${var.name}Function"
  api      = "${var.name}API"
  role     = "${var.name}Role"
  policy   = "${var.name}Policy"
  perm     = "${var.name}Permission"
  funclog  = "/aws/lambda/${local.func}"
  apilog   = "/aws/apigateway/${local.api}"
  envDebug = "${local.exe}Debug"
  envToken = "${local.exe}Token"
}

# GO111MODULE=on go get github.com/yi-jiayu/terraform-provider-telegram
provider "telegram" {
  bot_token = var.token
}
resource "telegram_bot_commands" "a" {
  commands = [{
    command     = "start"
    description = "— справка"
    }, {
    command     = "rules"
    description = "— правила"
    }, {
    command     = "hare"
    description = "слова… — задать слова для игры (через пробел, запятую или с новой строки)"
  }]
}
resource "telegram_bot_webhook" "a" {
  url             = aws_apigatewayv2_api.a.api_endpoint
  max_connections = 100
}

provider "archive" {
  version = "~> 1.3"
}
data "archive_file" "a" {
  type        = "zip"
  source_file = local.exe
  output_path = local.zip
}

provider "aws" {
  version = "~> 2.59"
  region  = var.region
}

resource "aws_apigatewayv2_api" "a" {
  name          = local.api
  protocol_type = "HTTP"
  target        = aws_lambda_function.a.arn
  route_key     = "POST /"
}
resource "aws_cloudwatch_log_group" "api" {
  # Since API logging is disabled by default this group is not automatically used.
  name              = local.apilog
  retention_in_days = 1
}
resource "aws_lambda_permission" "a" {
  statement_id  = local.perm
  function_name = aws_lambda_function.a.function_name
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.a.execution_arn}/*/*/*" # Any stage, method, resource.
}

resource "aws_lambda_function" "a" {
  function_name = local.func
  runtime       = "go1.x"
  handler       = local.exe
  memory_size   = 128 # MB, 128 + 64*x
  timeout       = 60  # seconds
  role          = aws_iam_role.a.arn

  filename         = data.archive_file.a.output_path
  source_code_hash = data.archive_file.a.output_base64sha256

  environment {
    variables = {
      (local.envDebug) = var.debug ? "1" : ""
      (local.envToken) = var.token
    }
  }
}
resource "aws_cloudwatch_log_group" "func" {
  # AWS Lambda automatically logs to the group with this name.
  name              = local.funclog
  retention_in_days = 1
}

resource "aws_iam_role_policy_attachment" "a" {
  role       = aws_iam_role.a.name
  policy_arn = aws_iam_policy.a.arn
}
resource "aws_iam_role" "a" {
  name               = local.role
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}
data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}
resource "aws_iam_policy" "a" {
  name   = local.policy
  policy = data.aws_iam_policy_document.a.json
}
data "aws_iam_policy_document" "a" {
  # Based on AWSLambdaBasicExecutionRole and AWSLambdaMicroserviceExecutionRole.
  statement {
    actions   = ["logs:CreateLogStream", "logs:PutLogEvents"]
    resources = [aws_cloudwatch_log_group.func.arn]
  }
  statement {
    actions   = ["dynamodb:UpdateItem"]
    resources = [aws_dynamodb_table.a.arn]
  }
}

resource "aws_dynamodb_table" "a" {
  name         = local.table
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "ID"
  attribute {
    type = "N"
    name = "ID"
  }
  ttl {
    enabled        = true
    attribute_name = "Expired"
  }
}
