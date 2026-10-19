from locust import FastHttpUser, constant, task


class ProductStressUser(FastHttpUser):
    wait_time = constant(0)

    @task
    def get_product(self):
        self.client.get(
            "/products/1",
            name="GET /products/:id",
            timeout=5
        )