
from locust import FastHttpUser, HttpUser, TaskSet, between, task


class ProductMixedTasks(TaskSet):

    # 80% GET
    @task(4)
    def get_product(self):
        self.client.get(
            "/products/1",
            name="GET /products/:id"
        )

    # 20% POST
    @task(1)
    def update_product(self):
        self.client.post(
            "/products/1/details",
            json={
                "product_id": 1,
                "sku": "LOCUST-MIXED",
                "manufacturer": "Apple",
                "category_id": 10,
                "weight": 250,
                "some_other_id": 100
            },
            name="POST /products/:id/details"
        )


class ProductMixedHttpUser(HttpUser):
    tasks = [ProductMixedTasks]
    wait_time = between(1, 2)


class ProductMixedFastHttpUser(FastHttpUser):
    tasks = [ProductMixedTasks]
    wait_time = between(1, 2)
