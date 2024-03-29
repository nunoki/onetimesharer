package server

// indexHTML returns the contents of the index.html page
func (serv server) indexHTML() string {
	return `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>One-time secret sharer</title>
	</head>
	<body>
		<div class="container">
			{{ if .ShareURL }}
	
				<div class="box">
					<textarea id="share_url" rows="4" readonly>{{ .ShareURL }}</textarea>
					<p class="center">Copy the URL and share it</p>
					<p class="center">
						<a href="/">
							Create another
						</a>
					</p>
				</div>
				<script>
					let share_url = document.getElementById('share_url');
					share_url.select();
					share_url.onclick = (function(){ this.select() });
				</script>
	
			{{ else if .SecretKey }}
	
				<div class="box center">
					<div id="secret_content">
						<p>
							When you click the button, the secret will be shown only once, and then
							deleted forever.
						</p>
						<button onclick="showSecret('{{ .SecretKey }}')">
							Show secret
						</button>
	
						<script>
							function showSecret(key) {
								let content = document.getElementById('secret_content');
								content.innerHTML = 'Loading...';
	
								let data = new FormData();
								data.append('key', key);
								fetch('/secret', {
									method: 'POST',
									body: data,
									'Content-type': 'application/x-www-form-urlencoded',
								})
									.then(content => content.json())
									.then(data => {
										if(!data.secret || !data.secret.length) {
											content.innerHTML = 'Something went wrong';
											content.classList.add('error');
											return
										}

										content.innerHTML = data.secret.
											replace(/[\u00A0-\u9999<>\&]/g, function(i) {
												return '&#'+i.charCodeAt(0)+';';
											});
										document.getElementById('create_another').removeAttribute('style');
									})
									.catch(error => {
										content.innerHTML = 'An error occurred';
										content.classList.add('error');
										console.error(error);
									});
							}
						</script>
					</div>
	
					<p class="center" id="create_another" style="display:none">
						<a href="/">
							Create new secret
						</a>
					</p>
				</div>
	
			{{ else if .ErrorMsg }}
	
				<div class="box error center">
					<p>{{ .ErrorMsg }}</p>
				</div>

				<p class="center">
					<a href="/">
						Create new secret
					</a>
				</p>

			{{ else }}
	
				<form action="" method="post">
					<input
						type="text"
						name="secret"
						placeholder="Secret to share one time"
						required
						autofocus
					/>
					<input
						type="text"
						name="signature"
					/>
					<button type="submit">
						Generate share URL
					</button>
				</form>
	
			{{ end }}
		</div>
	
		<style>
			html,
			body {
				padding:0; margin:0; width:100%; height:100%;
			}
			body {
				background: white; color: black;
			}
			* {
				font-size:1.2rem; box-sizing:border-box;
			}
			input,
			textarea {
				display:block; width:100%; padding:.6rem; word-break:break-all;
			}
			button {
				display:block; margin-top:.5rem; padding:.6rem; width:100%;
			}
			input,
			textarea,
			button,
			h4,
			p {
				margin-bottom:1rem;
			}
			.center {
				text-align:center;
			}
			.error {
				background-color:maroon;
				color: white;
			}
			.box {
				padding:1rem; width:100%; max-width:650px;
			}
			.container {
				display:flex; flex-direction: column; justify-content: center; 
				padding:1rem; align-items:center; height:100%;
			}
			input[name=signature] {
				display: none;
			}
			@media(prefers-color-scheme:dark) {
				body {
					background: #111;
					color: white;
				}
			}
		</style>
	</body>
	</html>
	`
}
