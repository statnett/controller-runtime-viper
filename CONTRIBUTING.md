# Submission Guidelines

## Submitting an Issue

Before you submit an issue, please search the issue tracker. An issue for your problem might already exist and the discussion might inform you of workarounds readily available.

We want to fix all the issues as soon as possible, but before fixing a bug, we need to reproduce and confirm it.
To reproduce bugs, we require that you provide minimal reproduction.
Having a minimal reproducible scenario gives us a wealth of important information without going back and forth to you with additional questions.

## Submitting a Pull Request (PR)

Before you submit your Pull Request (PR) consider the following guidelines:

1. Search [GitHub](https://github.com/statnett/controller-runtime-viper/pulls) for an open or closed PR that relates to your submission.
   You don't want to duplicate existing efforts.

2. Be sure that an issue describes the problem you're fixing, or documents the design for the feature you'd like to add.
   Discussing the design upfront helps to ensure that we're ready to accept your work.

3. [Fork](https://github.com/statnett/controller-runtime-viper) the repo.

4. In your forked repository, make your changes in a new git branch:

     ```shell
     git checkout -b my-fix-branch main
     ```

5. Create your patch, **including appropriate test cases**.

6. Run the unit tests and ensure that all tests pass.
    ```shell
     go test ./...
     ```

7. Run `golangci-lint` to catch any linter errors.
    ```shell
     golangci-lint run
     ```

8. Commit your changes using a descriptive commit message. We follow [Semantic release](https://github.com/semantic-release/semantic-release) to determine next semantic version number, generate a changelog and publish the release. Adherence to these conventions is necessary because release notes are automatically generated from these messages.

     ```shell
     git commit
     ```

9. Push your branch to GitHub:

    ```shell
    git push origin my-fix-branch
    ```

10. In GitHub, send a pull request to `main`.
