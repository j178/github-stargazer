# GitHub Stargazer

Receive notifications when someone starred your project.

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fj178%2Fgithub-stargazer&env=BARK_KEY&project-name=github-stargazer&repository-name=github-stargazer)

After deployment on Vercel, create a new webhook on the repositories you want to watch, and set the payload URL to something like `https://<project-name>.vercel.app/api`.

<img width="600" alt="image" src="https://user-images.githubusercontent.com/10510431/228465114-9732d9d3-c54f-4852-8e27-e9fc6fd7b660.png">

You only need to enable the `star` event type:

<img width="385" alt="image" src="https://user-images.githubusercontent.com/10510431/228465784-67183434-91f6-4f6b-92ed-b84fbf39a505.png">

