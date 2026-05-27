package com.poc.pedidos.repository;

import com.poc.pedidos.entity.OutboxEvent;
import org.springframework.data.jpa.repository.JpaRepository;
import java.util.UUID;

public interface OutboxRepository extends JpaRepository<OutboxEvent, UUID> {}
