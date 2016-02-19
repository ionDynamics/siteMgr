{{template "header" .}}
	<form class="form-signin sm-login" method="post" action="/login">

		<h2 class="form-signin-heading">Anmelden</h2>

		<div class="form-group">
			<label for="login-name" class="sr-only">Benutzername</label>
			<input type="text" id="login-name" name="login-name" class="form-control" placeholder="Benutzername" required autofocus>
		</div>

		<div class="form-group">
			<label for="login-pass" class="sr-only">Passwort</label>
			<input type="password" id="login-pass" name="login-pass" class="form-control" placeholder="Passwort" required>
		</div>

		<button class="btn btn-lg btn-primary btn-block" type="submit">Anmelden</button>
	</form>
{{template "footer" .}}