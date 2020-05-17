# AppEnv Generator

A custom [controller generator](https://book.kubebuilder.io/reference/controller-gen.html) to generate go functions.
The goal is to Have a simple interface to fetch environment variables from a custom resource.

This project was initiated from [BanzaiCloud blog post](https://banzaicloud.com/blog/generating-go-code/?utm_sq=ge2w5ug1pu) and the associated github repository [banzaicloud/go-code-generation-demo](https://github.com/banzaicloud/go-code-generation-demo).

## Install

```sh
make build BINARY='<full-destination-path>'
```

## How to use

1. Describe the Custom Resource Definition with markers

   ```go
   type SimpleSpec struct {
       // +appenv:key=MY_ENVIRONMENT_VARIABLE
       MySpecialConfig string `json:"mySpecialConfig"`
   }
   ```

2. Run the generator in your project

   ```sh
   ./bin/generator appenv ./apis/...
   ```

3. Use the new `GetApplicationEnvironments()` method in the controller.

   ```go
   myDeployment := &appsv1.Deployment{
       Spec: appsv1.DeploymentSpec{
           Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Env: myapp.GetApplicationEnvironments(context.TODO()),
                            // ...
                        },
                    },
                    // ...
                },
                // ...
           },
           // ...
       },
       // ...
   }
   ```
