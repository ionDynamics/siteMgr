{{define "html_header"}}<!DOCTYPE html>
<html lang="de">
	<head>
		<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon">
		<link rel="icon" href="/favicon.ico" type="image/x-icon">
		<meta charset="utf-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1">


		<link rel="stylesheet" href="/assets/css/bootstrap.min.css" type="text/css">
		<link rel="stylesheet" href="/assets/css/bootstrap-theme.min.css" type="text/css">
		<link rel="stylesheet" href="/assets/css/sitemgr.css" type="text/css">

		<script type="text/javascript" src="/assets/js/jquery-2.1.4.min.js"></script>
		<script type="text/javascript" src="/assets/js/bootstrap.min.js"></script>
		<script type="text/javascript" src="/assets/js/sitemgr.js"></script>
		<title>
			{{with .HtmlTitle}}
				{{.}} | SiteMgr
			{{else}}
				SiteMgr
			{{end}}
		</title>	
	</head>
	<body>
		{{template "navigation" .}}		
		<div class="container" id="main">
		{{with .FlashMessage}}
			<h2 class="flashmessage">{{.}}</h2>
		{{end}}
		<noscript>
			<h1 id="nojs" class="alert alert-danger">
				<span class="glyphicon glyphicon-warning-sign"></span>
				&nbsp;&nbsp;&nbsp;You have to enable JavaScript!&nbsp;&nbsp;&nbsp;
				<span class="glyphicon glyphicon-warning-sign"></span>
			</h1>
		</noscript>
{{end}}