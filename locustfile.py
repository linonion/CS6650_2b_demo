from locust import FastHttpUser, HttpUser, between, task


class ProductHttpUser(HttpUser):
    weight = 1
    wait_time = between(1, 2)

    @task
    def get_product(self):
        self.client.get("/products/1", name="GET /products/:id")


class ProductFastHttpUser(FastHttpUser):
    weight = 1
    wait_time = between(1, 2)

    @task
    def get_product(self):
        self.client.get("/products/1", name="GET /products/:id")