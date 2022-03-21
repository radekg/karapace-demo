CURRENT_DIR=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: clean
clean:
	-cd ${CURRENT_DIR}/.docker \
		&& docker-compose -f compose.yaml stop \
		&& echo y | docker-compose -f compose.yaml rm \
		&& docker volume rm \
			vol_kafka0_data \
			vol_kafka1_data \
			vol_kafka2_data \
			vol_zk1_data \
			vol_zk2_data \
			vol_zk3_data \
			vol_zk1_log \
			vol_zk2_log \
			vol_zk3_log

.PHONY: up
up:
	cd ${CURRENT_DIR}/.docker && docker-compose -f compose.yaml up

