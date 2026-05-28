package com.poc.pedidos.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.poc.pedidos.dto.CriarPedidoRequest;
import com.poc.pedidos.dto.PedidoResponse;
import com.poc.pedidos.entity.OutboxEvent;
import com.poc.pedidos.entity.Pedido;
import com.poc.pedidos.repository.OutboxRepository;
import com.poc.pedidos.repository.PedidoRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@Service
public class PedidoService {

    private static final Logger logger = LoggerFactory.getLogger(PedidoService.class);
    private static final String EVENTO_PEDIDO_CRIADO = "PedidoCriado";

    private final PedidoRepository pedidoRepository;
    private final OutboxRepository outboxRepository;
    private final ObjectMapper objectMapper;

    public PedidoService(PedidoRepository pedidoRepository,
                         OutboxRepository outboxRepository,
                         ObjectMapper objectMapper) {
        this.pedidoRepository = pedidoRepository;
        this.outboxRepository = outboxRepository;
        this.objectMapper = objectMapper;
    }

    /**
     * Cria o pedido e o evento na outbox na mesma transação.
     *
     * Se a gravação do evento falhar,
     * a gravação do pedido também é revertida, deste modo não fica um pedido
     * persistindo sem o respetivo evento PedidoCriado preparado para publicação.
     */
    @Transactional
    public PedidoResponse criarPedido(CriarPedidoRequest request) {
        Pedido pedido = new Pedido();
        pedido.setClienteId(request.getClienteId());
        pedido.setValor(request.getValor());

        Pedido pedidoGuardado = pedidoRepository.saveAndFlush(pedido);
        if (logger.isInfoEnabled()) {
            logger.info("Pedido criado: id={}, clienteId={}, valor={}",
                    pedidoGuardado.getId(), pedidoGuardado.getClienteId(), pedidoGuardado.getValor());
        }

        OutboxEvent evento = new OutboxEvent();
        evento.setTipo(EVENTO_PEDIDO_CRIADO);
        evento.setPayload(criarPayloadPedidoCriado(pedidoGuardado));

        OutboxEvent eventoGuardado = outboxRepository.save(evento);
        if (logger.isInfoEnabled()) {
            logger.info("Evento gravado na outbox: id={}, tipo={}, pedidoId={}",
                    eventoGuardado.getId(), eventoGuardado.getTipo(), pedidoGuardado.getId());
        }

        return PedidoResponse.from(pedidoGuardado);
    }

    @Transactional(readOnly = true)
    public PedidoResponse obterPedido(UUID id) {
        return pedidoRepository.findById(id)
                .map(PedidoResponse::from)
                .orElseThrow(() -> new PedidoNaoEncontradoException(id));
    }

    @Transactional(readOnly = true)
    public List<PedidoResponse> listarPedidosRecentes() {
        return pedidoRepository.findTop20ByOrderByCriadoEmDesc()
                .stream()
                .map(PedidoResponse::from)
                .toList();
    }

    private String criarPayloadPedidoCriado(Pedido pedido) {
        Map<String, Object> payload = Map.of(
                "pedidoId", pedido.getId().toString(),
                "clienteId", pedido.getClienteId().toString(),
                "valor", pedido.getValor(),
                "estado", pedido.getEstado(),
                "timestamp", OffsetDateTime.now().toString()
        );

        try {
            return objectMapper.writeValueAsString(payload);
        } catch (JsonProcessingException e) {
            throw new IllegalStateException("Não foi possível serializar o payload do evento PedidoCriado", e);
        }
    }
}
