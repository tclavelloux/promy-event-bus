FROM redis:7-alpine

# Enable append-only file persistence for data durability
# This ensures data survives container restarts
CMD ["redis-server", "--appendonly", "yes"]

