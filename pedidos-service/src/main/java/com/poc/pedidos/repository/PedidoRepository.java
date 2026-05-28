package com.poc.pedidos.repository;

import com.poc.pedidos.entity.Pedido;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.UUID;

public interface PedidoRepository extends JpaRepository<Pedido, UUID> {
    List<Pedido> findTop20ByOrderByCriadoEmDesc();
}
