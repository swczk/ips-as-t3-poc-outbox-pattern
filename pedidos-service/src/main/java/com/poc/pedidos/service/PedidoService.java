package com.poc.pedidos.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.poc.pedidos.dto.CriarPedidoRequest;
import com.poc.pedidos.dto.PedidoResponse;
import com.poc.pedidos.entity.OutboxEvent;
import com.poc.pedidos.entity.Pedido;
import com.poc.pedidos.repository.OutboxRepository;
import com.poc.pedidos.repository.PedidoRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.OffsetDateTime;
import java.util.Map;

@Service
public class PedidoService {

    private final PedidoRepository pedidoRepository;
    private final OutboxRepository outboxRepository;
    private final ObjectMapper objectMapper;

    public PedidoService(PedidoRepository pedidoRepository,
                         OutboxRepository outboxRepository,
                         ObjectMapper objectMapper) {
        this.pedidoRepository = pedidoRepository;
        this.outboxRepository = outboxRepository;
        this.objectMapper     = objectMapper;
    }

    @Transactional
    public PedidoResponse criarPedido(CriarPedidoRequest request) throws Exception {
        Pedido pedido = new Pedido();
        pedido.setClienteId(request.getClienteId());
        pedido.setValor(request.getValor());
        pedidoRepository.save(pedido);

        Map<String, Object> payload = Map.of(
            "pedidoId",  pedido.getId().toString(),
            "clienteId", pedido.getClienteId().toString(),
            "valor",     pedido.getValor(),
            "timestamp", OffsetDateTime.now().toString()
        );

        OutboxEvent evento = new OutboxEvent();
        evento.setTipo("PedidoCriado");
        evento.setPayload(objectMapper.writeValueAsString(payload));
        outboxRepository.save(evento);

        return PedidoResponse.from(pedido);
    }
}
